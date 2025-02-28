package models

type Meta struct {
	Default  Presentation            `json:"default"`
	Elements map[string]Presentation `json:"elements"`
}

type Presentation struct {
	Entity string `json:"entity"`
	Icon   string `json:"icon"`
}
