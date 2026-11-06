module github.com/msgmate-io/open-chat-go-community-coder-integration

go 1.26.0

require github.com/msgmate-io/go-integration-interface v0.0.0

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/msgmate-io/go-tool-interface v0.0.0 // indirect
	github.com/urfave/cli/v3 v3.11.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5
	gorm.io/gorm v1.25.12 // indirect
)

replace github.com/msgmate-io/go-integration-interface => ../../go_integration_interface

replace github.com/msgmate-io/go-tool-interface => ../../go_tool_interface
