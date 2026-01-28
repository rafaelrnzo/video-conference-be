package models

import (
	"video-conference-be/internal/domain/group"
	"video-conference-be/internal/domain/poll"
	"video-conference-be/internal/domain/record"
	"video-conference-be/internal/domain/room"
	"video-conference-be/internal/domain/role"
	"video-conference-be/internal/domain/user"
)

var Models = []interface{}{
	&user.User{},
	&room.Room{},
	&record.Record{},
	&group.Group{},
	&role.Role{},
	&role.Permission{},
	&poll.Poll{},
}
