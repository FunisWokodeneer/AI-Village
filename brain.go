package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type decisionRequest struct {
	Model string `json:"model"`
	Messages []map[string]string `json:"messages"`
	Temperature float64 `json:"temperature"`
	MaxTokens int `json:"max_tokens"`
}
type decisionResponse struct {
	Choices []struct { Message struct { Content string `json:"content"` } `json:"message"` } `json:"choices"`
}

func (g *VillageGame) requestDecision(a *Resident) {
	a.Thinking = true
	defer func(){ a.Thinking=false }()
	key:=os.Getenv("AI_VILLAGE_API_KEY")
	if key=="" { key=os.Getenv("NIM_API_KEY") }
	if key=="" { return }
	endpoint:=os.Getenv("AI_VILLAGE_ENDPOINT")
	if endpoint=="" { endpoint="https://integrate.api.nvidia.com/v1/chat/completions" }
	model:=os.Getenv("AI_VILLAGE_MODEL")
	if model=="" { model="meta/llama-3.1-70b-instruct" }
	system:=fmt.Sprintf("You are %s, an AI resident in AGS AI Village. Role: %s. Personality: %s. Choose one short, safe, ordinary activity and a destination from this list: AGS Commons, Research House, Workshop, Creative Studio, Garden Cafe. Reply with JSON only: {\"activity\":\"...\",\"destination\":\"...\"}.",a.Name,a.Role,a.Personality)
	body:=decisionRequest{Model:model,Temperature:0.4,MaxTokens:120,Messages:[]map[string]string{{"role":"system","content":system},{"role":"user","content":"Current activity: "+a.CurrentAction+". Decide what to do next."}}}
	payload,err:=json.Marshal(body); if err!=nil{return}
	req,err:=http.NewRequest(http.MethodPost,endpoint,bytes.NewReader(payload)); if err!=nil{return}
	req.Header.Set("Content-Type","application/json"); req.Header.Set("Authorization","Bearer "+key)
	client:=http.Client{Timeout:20*time.Second}
	resp,err:=client.Do(req); if err!=nil{return}; defer resp.Body.Close()
	var result decisionResponse
	if json.NewDecoder(resp.Body).Decode(&result)!=nil || resp.StatusCode<200 || resp.StatusCode>=300 || len(result.Choices)==0{return}
	var choice struct{Activity string `json:"activity"`; Destination string `json:"destination"`}
	content:=strings.TrimSpace(result.Choices[0].Message.Content)
	content=strings.TrimPrefix(content,"```json"); content=strings.TrimSuffix(content,"```")
	if json.Unmarshal([]byte(strings.TrimSpace(content)),&choice)!=nil{return}
	targets:=map[string]Point{"AGS Commons":{550,335},"Research House":{190,260},"Workshop":{900,275},"Creative Studio":{915,555},"Garden Cafe":{200,570}}
	if p,ok:=targets[choice.Destination];ok { a.Target=p }
	if choice.Activity!="" { a.CurrentAction=choice.Activity }
}
