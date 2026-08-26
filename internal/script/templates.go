package script

import (
	"fmt"

	"example.com/materialconsole/internal/model"
)

func designHost(project model.Project) model.Script {
	return model.Script{ID: project.ID + "-host", ProjectID: project.ID, Kind: "host", Title: "主持人口播", Lines: []string{fmt.Sprintf("欢迎来到%s新品发布。", project.Product), "今天我们从真实场景出发，看看它如何解决日常问题。", project.Slogan}, Duration: 30, Status: model.StatusNew, Version: 1}
}

func designProduct(project model.Project) model.Script {
	lines := []string{"镜头切入产品三维占位场景。", "展示核心结构与材质细节。"}
	for _, point := range project.SellingPoints {
		lines = append(lines, "重点卖点："+point)
	}
	return model.Script{ID: project.ID + "-product", ProjectID: project.ID, Kind: "product-3d", Title: "产品三维展示段", Lines: lines, Duration: 45, Status: model.StatusNew, Version: 1}
}

func designRecall(project model.Project) model.Script {
	channel := "全渠道"
	if len(project.Channels) > 0 {
		channel = project.Channels[0]
	}
	return model.Script{ID: project.ID + "-recall", ProjectID: project.ID, Kind: "recall", Title: "片尾召回", Lines: []string{"记住今天的新品，也记住它为你节省的每一步。", "现在前往" + channel + "了解完整信息。"}, Duration: 15, Status: model.StatusNew, Version: 1}
}
