package vo

type WhosTakingLeaveList map[string][]WhosTakingLeaveElements

type WhosTakingLeaveElements struct {
	// Leave's ID
	ID string `json:"id,omitempty"`
	// Leave Type
	Type string `json:"type,omitempty"`
	// Employee's avatar taking leave that day
	Avatar string `json:"avatar,omitempty"`
	// Employee's fullName taking leave that day
	FullName string `json:"fullName,omitempty"`
	// Employee's role taking leave that day
	Role string `json:"role,omitempty"`
}
