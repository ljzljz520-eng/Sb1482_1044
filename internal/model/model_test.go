package model

import "testing"

func TestProjectLifecycleHelpers(t *testing.T) {
	project := NewProject("p-1", "新品素材", "Aurora One")
	project.SellingPoints = []string{"轻量", "长续航"}
	project.Channels = []string{"video", "store"}
	project.Slogan = "把灵感带上路"
	project.ReferenceVideo = "ref://aurora"
	project.Materials = []MaterialRecord{{ID: "I-109", Title: "主视觉"}}
	project.Scripts = []Script{{Kind: "host"}, {Kind: "product-3d"}, {Kind: "recall"}}
	project.Timeline = []Timeline{{Seconds: 10}, {Seconds: 20}, {Seconds: 5}}
	if err := project.Valid(); err != nil {
		t.Fatal(err)
	}
	if !project.HasMaterial("I-109") || project.TotalSeconds() != 35 {
		t.Fatalf("unexpected project helpers: %+v", project)
	}
	if !project.ReadyForReview() {
		t.Fatal("project should be reviewable")
	}
	clone := project.Clone()
	clone.SellingPoints[0] = "独立"
	if project.SellingPoints[0] == clone.SellingPoints[0] {
		t.Fatal("clone shares selling points")
	}
}
