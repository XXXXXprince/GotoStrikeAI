package handler

import (
    "testing"
)

// smoke test for the internal wecom encryption/decryption helpers.  the
// functions are intentionally unexported because they're implementation
// detail for the robot handler, but we can still write a simple unit test
// in the same package to verify that round trips work with real values from
// our configuration.

func TestWecomEncryptDecrypt(t *testing.T) {
    // these values are pulled from the example config that is used in the
    // workspace; the encode key must be the 43-character string (base64
    // encoded AES key) and the corpID should match the ToUserName that we
    // receive in callbacks.
    encodingKey := "TupAQh8HeOFYLrRhH8xc3wvJeds6nu1Xr0hn1Lfy1Gh"
    corpID := "wwf37149f596dd5638"

    // build a sample XML payload exactly how sendWecomReply constructs it.
    msg := `<xml>` +
        `<ToUserName><![CDATA[toUser]]></ToUserName>` +
        `<FromUserName><![CDATA[fromUser]]></FromUserName>` +
        `<CreateTime>12345</CreateTime>` +
        `<MsgType><![CDATA[text]]></MsgType>` +
        `<Content><![CDATA[test]]></Content>` +
        `<AgentID><![CDATA[1000002]]></AgentID>` +
        `</xml>`

    enc, err := wecomEncrypt(encodingKey, msg, corpID)
    if err != nil {
        t.Fatalf("encrypt error: %v", err)
    }

    dec, err := wecomDecrypt(encodingKey, enc)
    if err != nil {
        t.Fatalf("decrypt error: %v", err)
    }

    if string(dec) != msg {
        t.Fatalf("round trip mismatch; got %q", string(dec))
    }
}
