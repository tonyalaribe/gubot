package middleware

import (
	"github.com/ArthurHlt/gubot/robot"
)

var authorizeConfig AuthorizeConfig

func init() {
	robot.GetConfig(&authorizeConfig)
}

type AuthorizeConfig struct {
	AccessControl []AccessControl `cloud:"auth_access_control"`
	Groups        []Group         `cloud:"auth_groups"`
}

func (g AuthorizeConfig) GetAccessControl(scriptName string) AccessControl {
	for _, ac := range g.AccessControl {
		if ac.Name == scriptName {
			return ac
		}
	}
	return AccessControl{}
}

type Groups []Group

func (g Groups) GetGroup(groupName string) Group {
	for _, group := range g {
		if group.Name == groupName {
			return group
		}
	}
	return Group{}
}

type Group struct {
	Name  string
	Users []string
}

func (g Group) HasAccess(currentUser string) bool {
	for _, user := range g.Users {
		if user == currentUser {
			return true
		}
	}
	return false
}

type AccessControl struct {
	Name   string `cloud:"name"`
	Users  []string
	Groups []string
}

func (g AccessControl) HasAccess(currentUser string, groups Groups) bool {
	for _, user := range g.Users {
		if user == currentUser {
			return true
		}
	}
	for _, groupName := range g.Groups {
		group := groups.GetGroup(groupName)
		if group.Name == "" {
			continue
		}
		if group.HasAccess(currentUser) {
			return true
		}
	}
	return false
}

type AuthorizeMiddleware struct{}

func (AuthorizeMiddleware) ScriptMiddleware(script robot.Script, next robot.EnvelopHandler) robot.EnvelopHandler {
	return func(envelop robot.Envelop, submatch [][]string) ([]string, error) {
		ac := authorizeConfig.GetAccessControl(script.Name)
		groups := authorizeConfig.Groups
		if ac.Name == "" {
			return next(envelop, submatch)
		}
		if ac.HasAccess(envelop.User.Name, groups) || ac.HasAccess(envelop.User.Id, groups) {
			return next(envelop, submatch)
		}
		return []string{}, nil
	}
}

func (AuthorizeMiddleware) CommandMiddleware(command robot.SlashCommand, next robot.CommandHandler) robot.CommandHandler {
	return func(envelop robot.Envelop) (string, error) {
		ac := authorizeConfig.GetAccessControl(command.Trigger)
		groups := authorizeConfig.Groups
		if ac.Name == "" {
			return next(envelop)
		}
		if ac.HasAccess(envelop.User.Name, groups) || ac.HasAccess(envelop.User.Id, groups) {
			return next(envelop)
		}
		return "", nil
	}
}
