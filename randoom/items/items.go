package items

type Item struct {
	Content string  `json:"content" yaml:"content"`
	Count   uint64  `json:"count" yaml:"count"`
	Ratio   float32 `json:"ratio" yaml:"ratio"`
}

type List struct {
	Label string `json:"label" yaml:"label"`
	Items []Item `json:"items" yaml:"items"`
}
