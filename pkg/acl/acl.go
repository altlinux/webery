package acl

type Member struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Leader bool   `json:"leader"`
}

type ACL struct {
	Repo    string   `json:"repo"`
	Name    string   `json:"name"`
	Members []Member `json:"members"`
}
