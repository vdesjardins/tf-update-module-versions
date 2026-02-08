package report

var outputWriter writer = &defaultWriter{}

// SetOutput configures a global output writer used for diff rendering.
func SetOutput(diffTool string) error {
	if diffTool == "" {
		outputWriter = &defaultWriter{}
		return nil
	}

	toolWriter, err := newToolWriter(diffTool)
	if err != nil {
		return err
	}

	outputWriter = toolWriter
	return nil
}

// RenderOutput passes input through the configured output writer.
func RenderOutput(input string) (string, error) {
	return outputWriter.Write(input)
}
