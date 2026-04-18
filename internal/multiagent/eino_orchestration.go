package multiagent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// PlanExecuteRootArgs 构建 Eino adk/prebuilt/planexecute 根 Agent 所需参数。
type PlanExecuteRootArgs struct {
	MainToolCallingModel *openai.ChatModel
	ExecModel            *openai.ChatModel
	OrchInstruction      string
	ToolsCfg             adk.ToolsConfig
	ExecMaxIter          int
	LoopMaxIter          int
}

// NewPlanExecuteRoot 返回 plan → execute → replan 预置编排根节点（与 Deep / Supervisor 并列）。
func NewPlanExecuteRoot(ctx context.Context, a *PlanExecuteRootArgs) (adk.ResumableAgent, error) {
	if a == nil {
		return nil, fmt.Errorf("plan_execute: args 为空")
	}
	if a.MainToolCallingModel == nil || a.ExecModel == nil {
		return nil, fmt.Errorf("plan_execute: 模型为空")
	}
	tcm, ok := interface{}(a.MainToolCallingModel).(model.ToolCallingChatModel)
	if !ok {
		return nil, fmt.Errorf("plan_execute: 主模型需实现 ToolCallingChatModel")
	}
	planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: tcm,
	})
	if err != nil {
		return nil, fmt.Errorf("plan_execute planner: %w", err)
	}
	replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: tcm,
	})
	if err != nil {
		return nil, fmt.Errorf("plan_execute replanner: %w", err)
	}
	executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model:         a.ExecModel,
		ToolsConfig:   a.ToolsCfg,
		MaxIterations: a.ExecMaxIter,
		GenInputFn:    planExecuteExecutorGenInput(a.OrchInstruction),
	})
	if err != nil {
		return nil, fmt.Errorf("plan_execute executor: %w", err)
	}
	loopMax := a.LoopMaxIter
	if loopMax <= 0 {
		loopMax = 10
	}
	return planexecute.New(ctx, &planexecute.Config{
		Planner:       planner,
		Executor:      executor,
		Replanner:     replanner,
		MaxIterations: loopMax,
	})
}

func planExecuteExecutorGenInput(orchInstruction string) planexecute.GenModelInputFn {
	oi := strings.TrimSpace(orchInstruction)
	return func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
		planContent, err := in.Plan.MarshalJSON()
		if err != nil {
			return nil, err
		}
		userMsgs, err := planexecute.ExecutorPrompt.Format(ctx, map[string]any{
			"input":          planExecuteFormatInput(in.UserInput),
			"plan":           string(planContent),
			"executed_steps": planExecuteFormatExecutedSteps(in.ExecutedSteps),
			"step":           in.Plan.FirstStep(),
		})
		if err != nil {
			return nil, err
		}
		if oi != "" {
			userMsgs = append([]adk.Message{schema.SystemMessage(oi)}, userMsgs...)
		}
		return userMsgs, nil
	}
}

func planExecuteFormatInput(input []adk.Message) string {
	var sb strings.Builder
	for _, msg := range input {
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func planExecuteFormatExecutedSteps(results []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for _, result := range results {
		sb.WriteString(fmt.Sprintf("Step: %s\nResult: %s\n\n", result.Step, result.Result))
	}
	return sb.String()
}

// planExecuteStreamsMainAssistant 将规划/执行/重规划各阶段助手流式输出映射到主对话区。
func planExecuteStreamsMainAssistant(agent string) bool {
	if agent == "" {
		return true
	}
	switch agent {
	case "planner", "executor", "replanner", "execute_replan", "plan_execute_replan":
		return true
	default:
		return false
	}
}

func planExecuteEinoRoleTag(agent string) string {
	_ = agent
	return "orchestrator"
}
