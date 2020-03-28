package model

// Onboard a request to onboard is stored in here, and all the tasks associated are created and also stored here
type Onboard struct {
	ManagerEmail  string  `json:"managerEmail"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	StartDate     string  `json:"startDate"`
	BeforeJoining []*Task `json:"beforeJoining"`
	AfterJoining  []*Task `json:"afterJoining"`
	Systems       []*Task `json:"systems"`
}

func NewOnboard() *Onboard {
	return &Onboard{
		BeforeJoining: []*Task{},
		AfterJoining:  []*Task{},
		Systems:       []*Task{},
	}
}

// Task A task to be done
type Task struct {
	Name          string `json:"name"`
	AssigneeEmail string `json:"assigneeEmail"`
}
