package global

var (
	// Auto 标志是否开启自动提交
	Auto = false

	// ODBStatus 标识ODB功能是否开启，默认开启
	ODBStatus int64 = 1

	// AOFStatus 标识AOF功能是否开启，默认不开启
	AOFStatus int64 = 0
)
