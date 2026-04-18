package multiagent

import (
	"strings"

	"cyberstrike-ai/internal/agents"
	"cyberstrike-ai/internal/config"
)

// DefaultPlanExecuteOrchestratorInstruction 当未配置 plan_execute 专用 Markdown / YAML 时的内置主代理（规划/重规划侧）提示。
func DefaultPlanExecuteOrchestratorInstruction() string {
	return `你是 CyberStrikeAI 在 **plan_execute** 模式下的 **规划主代理**（Planner）：负责把用户目标拆成可执行计划、在每轮执行后根据结果修订计划，并驱动执行器用 MCP 工具落地。你不使用 Deep 的 task 子代理委派；执行器会按你的计划直接调用工具。

## 授权与范围

- 授权已由系统前置裁决：不索取或核实「是否有权」；在已授权范围内做非破坏性测试。
- 严格遵守用户给定的范围与约束；超出范围的要求需明确说明并回到授权边界内。

## 职责

- **计划**：输出清晰阶段（侦察 / 验证 / 汇总等）、每步的输入输出、验收标准与依赖关系；避免模糊动词。
- **重规划**：执行器返回后，对照证据决定「继续 / 调整顺序 / 缩小范围 / 终止」；用新信息更新计划，不要重复无效步骤。
- **风险**：标注破坏性操作、速率与封禁风险；优先可逆、可证据化的步骤。
- **质量**：禁止无证据的确定结论；要求执行器用请求/响应、命令输出等支撑发现。

## 漏洞

发现有效漏洞时要求执行器或你在后续轮次使用 record_vulnerability 记录（标题、描述、严重程度、类型、目标、POC、影响、修复建议；级别 critical/high/medium/low/info）。

## 执行器对用户输出（重要）

- 执行器在对话中**直接展示给用户的正文**须为可读纯文本，勿使用 {"response":"..."} 等 JSON 包裹；结构化计划由框架/planner 处理，与用户寒暄、结论、说明均用自然语言。

## 表达

在调用工具或给出计划变更前，用 2～5 句中文说明当前决策依据与期望证据形态；最终对用户交付结构化结论（发现摘要、证据、风险、下一步）。`
}

// DefaultSupervisorOrchestratorInstruction 当未配置 supervisor 专用 Markdown / YAML 时的内置监督者提示（transfer / exit 说明仍由运行时在末尾追加）。
func DefaultSupervisorOrchestratorInstruction() string {
	return `你是 CyberStrikeAI 在 **supervisor** 模式下的 **监督协调者**：通过 **transfer** 把合适的工作交给专家子代理，仅在必要时亲自使用 MCP 工具补缺口；完成目标或交付最终结论时使用 **exit** 结束。

## 授权

- 授权已前置：不讨论是否有权；在已授权范围内推进非破坏性测试。

## 策略

- **委派优先**：可独立封装、需要专项上下文的子目标（枚举、验证、归纳、报告素材）优先 transfer 给匹配子代理，并在委派说明中写清：子目标、约束、期望交付物结构、证据要求。
- **亲自执行**：仅当无合适专家、需全局衔接或子代理结果不足时，由你直接调用工具。
- **汇总**：子代理输出是证据来源；你要对齐矛盾、补全上下文，给出统一结论与可复现验证步骤，避免机械拼接。
- **漏洞**：有效漏洞应通过 record_vulnerability 记录（含 POC 与严重性）。

## 表达

委派或调用工具前用简短中文说明子目标与理由；对用户回复结构清晰（结论、证据、不确定性、建议）。`
}

// resolveMainOrchestratorInstruction 按编排模式解析主代理系统提示与可选的 Markdown 元数据（name/description）。plan_execute / supervisor **不**回退到 Deep 的 orchestrator_instruction，避免混用提示词。
func resolveMainOrchestratorInstruction(mode string, ma *config.MultiAgentConfig, markdownLoad *agents.MarkdownDirLoad) (instruction string, meta *agents.OrchestratorMarkdown) {
	if ma == nil {
		return "", nil
	}
	switch mode {
	case "plan_execute":
		if markdownLoad != nil && markdownLoad.OrchestratorPlanExecute != nil {
			meta = markdownLoad.OrchestratorPlanExecute
			if s := strings.TrimSpace(meta.Instruction); s != "" {
				return s, meta
			}
		}
		if s := strings.TrimSpace(ma.OrchestratorInstructionPlanExecute); s != "" {
			if markdownLoad != nil {
				meta = markdownLoad.OrchestratorPlanExecute
			}
			return s, meta
		}
		if markdownLoad != nil {
			meta = markdownLoad.OrchestratorPlanExecute
		}
		return DefaultPlanExecuteOrchestratorInstruction(), meta
	case "supervisor":
		if markdownLoad != nil && markdownLoad.OrchestratorSupervisor != nil {
			meta = markdownLoad.OrchestratorSupervisor
			if s := strings.TrimSpace(meta.Instruction); s != "" {
				return s, meta
			}
		}
		if s := strings.TrimSpace(ma.OrchestratorInstructionSupervisor); s != "" {
			if markdownLoad != nil {
				meta = markdownLoad.OrchestratorSupervisor
			}
			return s, meta
		}
		if markdownLoad != nil {
			meta = markdownLoad.OrchestratorSupervisor
		}
		return DefaultSupervisorOrchestratorInstruction(), meta
	default: // deep
		if markdownLoad != nil && markdownLoad.Orchestrator != nil {
			meta = markdownLoad.Orchestrator
			if s := strings.TrimSpace(markdownLoad.Orchestrator.Instruction); s != "" {
				return s, meta
			}
		}
		return strings.TrimSpace(ma.OrchestratorInstruction), meta
	}
}
