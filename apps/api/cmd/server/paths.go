package main

func templatePath(name string) string {
	return cfg.TemplatePath(name)
}

func staticPath(name string) string {
	return cfg.StaticPath(name)
}
