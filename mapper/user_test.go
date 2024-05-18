package mapper

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"
)

func TestUserMapper(t *testing.T) {
	mapper := NewUserMapper()

	user := &model.User{
		ID:       1,
		Email:    "test@localhost",
		Password: "$ecret"}

	entity := mapper.MapUserModelToConfig(user)
	assert.NotNil(t, entity)

	assert.Equal(t, user.ID, entity.ID)
	assert.Equal(t, user.GetEmail(), entity.GetEmail())
	assert.Equal(t, user.GetPassword(), entity.GetPassword())

	model := mapper.MapUserConfigToModel(entity)
	assert.Equal(t, entity.ID, model.GetID())
	assert.Equal(t, entity.GetEmail(), model.GetEmail())
	assert.Equal(t, entity.GetPassword(), model.GetPassword())
}
