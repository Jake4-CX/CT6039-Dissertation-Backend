package structs

import (
	"encoding/json"
	"errors"
)

type TreeNode struct {
	Name       string      `json:"name" binding:"required"`
	Type       string      `json:"type" binding:"required"`
	Data       NodeData    `json:"data" binding:"required"`
	Children   []TreeNode  `json:"children" binding:"required"`
	Conditions *Conditions `json:"conditions"`
}

type Conditions struct {
	TrueChildren  []TreeNode `json:"trueChildren"`
	FalseChildren []TreeNode `json:"falseChildren"`
}

type NodeData interface {
	IsNodeData()
}

type GetRequestNode struct {
	Label string `json:"label" binding:"required"`
	URL   string `json:"url" binding:"required"`
}

type PostRequestNode struct {
	Label string `json:"label" binding:"required"`
	URL   string `json:"url" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type IfConditionNode struct {
	Label     string `json:"label" binding:"required"`
	Field     string `json:"field" binding:"required"`
	Condition string `json:"condition" binding:"required"`
	Value     string `json:"value" binding:"required"`
}

type StartNode struct {
	Label string `json:"label" binding:"required"`
}

type StopNode struct {
	Label string `json:"label" binding:"required"`
}

func (*GetRequestNode) IsNodeData()  {}
func (*PostRequestNode) IsNodeData() {}
func (*IfConditionNode) IsNodeData() {}
func (*StartNode) IsNodeData()       {}
func (*StopNode) IsNodeData()        {}

func (t *TreeNode) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name       string          `json:"name"`
		Type       string          `json:"type"`
		Data       json.RawMessage `json:"data"`
		Children   []TreeNode      `json:"children"`
		Conditions *Conditions     `json:"conditions"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var nodeData NodeData
	switch raw.Type {
	case "getRequest":
		nodeData = &GetRequestNode{}
	case "postRequest":
		nodeData = &PostRequestNode{}
	case "ifCondition":
		nodeData = &IfConditionNode{}
	case "startNode":
		nodeData = &StartNode{}
	case "stopNode":
		nodeData = &StopNode{}
	default:
		return errors.New("unknown type")
	}

	if err := json.Unmarshal(raw.Data, nodeData); err != nil {
		return err
	}

	// Populate the TreeNode fields.
	t.Name = raw.Name
	t.Type = raw.Type
	t.Data = nodeData
	t.Children = raw.Children
	t.Conditions = raw.Conditions
	return nil
}

// ReactFlow structs
type ReactFlow struct {
	Edges    []FlowEdge `json:"edges"`
	Nodes    []FlowNode `json:"nodes"`
	Viewport Viewport   `json:"viewport"`
}

type FlowEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`
}

type FlowNode struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
	PositionAbsolute struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"PositionAbsolute"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	Connectable bool    `json:"connectable"`
}

type Viewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

type UpdateTestPlanRequest struct {
	TestPlan  []TreeNode `json:"testPlan"`
	ReactFlow ReactFlow  `json:"reactFlow"`
}
