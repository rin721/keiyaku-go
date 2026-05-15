package cli

import (
	survey "github.com/AlecAivazis/survey/v2"
)

// SurveyPrompter 是基于 Survey 的默认交互实现。
type SurveyPrompter struct{}

// AskOne 调用 Survey 的单题询问。
func (SurveyPrompter) AskOne(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
	return survey.AskOne(prompt, response, opts...)
}

// AskString 读取字符串输入。
func (u *UI) AskString(message, defaultValue string) (string, error) {
	if u == nil || !u.interactive {
		return "", NewError(ErrorKindPrompt, OperationPrompt, "无法读取交互输入", ErrNonInteractive)
	}
	value := defaultValue
	prompt := &survey.Input{Message: message, Default: defaultValue}
	if err := u.prompter.AskOne(prompt, &value); err != nil {
		return "", NewError(ErrorKindPrompt, OperationPrompt, "读取字符串输入失败", err)
	}
	return value, nil
}

// AskConfirm 读取确认输入。
func (u *UI) AskConfirm(message string, defaultValue bool) (bool, error) {
	if u == nil || !u.interactive {
		return false, NewError(ErrorKindPrompt, OperationPrompt, "无法读取交互确认", ErrNonInteractive)
	}
	value := defaultValue
	prompt := &survey.Confirm{Message: message, Default: defaultValue}
	if err := u.prompter.AskOne(prompt, &value); err != nil {
		return false, NewError(ErrorKindPrompt, OperationPrompt, "读取确认输入失败", err)
	}
	return value, nil
}

// AskSelect 读取单选输入。
func (u *UI) AskSelect(message string, options []string, defaultValue string) (string, error) {
	if u == nil || !u.interactive {
		return "", NewError(ErrorKindPrompt, OperationPrompt, "无法读取交互选择", ErrNonInteractive)
	}
	value := defaultValue
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Default: defaultValue,
	}
	if err := u.prompter.AskOne(prompt, &value); err != nil {
		return "", NewError(ErrorKindPrompt, OperationPrompt, "读取选择输入失败", err)
	}
	return value, nil
}
