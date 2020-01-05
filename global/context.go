package global

type Context struct {
	WorkerNum int
	SiteName  string
}

type SiteConfig struct {
	TotalPageReg     string
	NextPageReg      string
	SeederReg        string
	TitleReg         string
	CategoryReg      string
	TagReg           string
	PostTimeReg      string
	PleasureReg      string
	ListUrl          []string
	PageUrlGenerator func(originUrl string, id string, num int) string
	IdGetter         func(url string) string
}

type Configs map[string]SiteConfig
