package handler

import "fmt"

type PanelUpdatePlanInfo struct {
	ContainerName string
	ImageName     string
	PullImageName string
	Channel       string
	MirrorHost    string
	RegistryURL   string
}

type PanelUpdateStatusInfo struct {
	Status        string
	Phase         string
	Message       string
	Error         string
	ContainerName string
	ImageName     string
	PullImageName string
	MirrorHost    string
	RegistryURL   string
}

func BuildPanelUpdatePlanInfo() (PanelUpdatePlanInfo, error) {
	plan, err := buildPanelUpdatePlan()
	if err != nil {
		return PanelUpdatePlanInfo{}, err
	}

	return PanelUpdatePlanInfo{
		ContainerName: plan.ContainerName,
		ImageName:     plan.ImageName,
		PullImageName: plan.PullImageName,
		Channel:       plan.Channel,
		MirrorHost:    plan.MirrorHost,
		RegistryURL:   plan.RegistryURL,
	}, nil
}

func ExecutePanelUpdateForCLI() (PanelUpdateStatusInfo, error) {
	plan, err := buildPanelUpdatePlan()
	if err != nil {
		return PanelUpdateStatusInfo{}, err
	}

	executePanelUpdate(plan)

	snapshot := panelUpdater.snapshotCopy()
	status := PanelUpdateStatusInfo{
		Status:        snapshot.Status,
		Phase:         snapshot.Phase,
		Message:       snapshot.Message,
		Error:         snapshot.Error,
		ContainerName: snapshot.ContainerName,
		ImageName:     snapshot.ImageName,
		PullImageName: snapshot.PullImageName,
		MirrorHost:    snapshot.MirrorHost,
		RegistryURL:   snapshot.RegistryURL,
	}

	if snapshot.Status == "failed" {
		if snapshot.Error != "" {
			return status, fmt.Errorf("%s", snapshot.Error)
		}
		return status, fmt.Errorf("%s", snapshot.Message)
	}

	return status, nil
}
