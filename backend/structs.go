package backend

type ConvertFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type Progress struct {
	Percentage  int          `json:"percentage"`
	CurrentFile *ConvertFile `json:"currentFile"`
	Running     bool         `json:"running"`
	Error       string       `json:"error"`
}

type ProcessData struct {
	Files        []ConvertFile `json:"files"`
	Transparent  int           `json:"transparent"`
	Size         int           `json:"size"`
	Watermark    string        `json:"watermark"`
	Prefix       string        `json:"prefix"`
	Position     string        `json:"position"`
	OutputFolder string        `json:"outputFolder"`
}
