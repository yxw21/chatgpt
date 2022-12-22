package chatgpt

type Asset struct {
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type Release struct {
	Assets []Asset `json:"assets"`
}
