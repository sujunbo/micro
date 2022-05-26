package db

type Options struct {
	Driver     string `default:"mysql"`
	DataSource string
	DBName     string
	UserName   string
	Password   string
	Host       string
	Port       int

	MaxIdleConns int
	MaxOpenConns int
}
