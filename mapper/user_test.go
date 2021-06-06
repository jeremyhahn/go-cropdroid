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

	entity := mapper.MapUserModelToEntity(user)
	assert.NotNil(t, entity)

	assert.Equal(t, user.GetID(), entity.GetID())
	assert.Equal(t, user.GetEmail(), entity.GetEmail())
	assert.Equal(t, user.GetPassword(), entity.GetPassword())

	model := mapper.MapUserEntityToModel(entity)
	assert.Equal(t, entity.GetID(), model.GetID())
	assert.Equal(t, entity.GetEmail(), model.GetEmail())
	assert.Equal(t, entity.GetPassword(), model.GetPassword())
}
