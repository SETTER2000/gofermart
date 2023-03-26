package repo

import (
	"context"
	"github.com/SETTER2000/gofermart/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockShorturlRepo struct {
	mock.Mock
}

func (mock *MockShorturlRepo) Post(context.Context, *entity.Gofermart) error {
	args := mock.Called()
	return args.Error(1)
}
func (mock *MockShorturlRepo) Put(context.Context, *entity.Gofermart) error {
	args := mock.Called()
	return args.Error(1)
}
func (mock *MockShorturlRepo) Get(context.Context, *entity.Gofermart) (*entity.Gofermart, error) {
	args := mock.Called()
	result := args.Get(1)
	return result.(*entity.Gofermart), args.Error(1)
}
func (mock *MockShorturlRepo) GetAll(context.Context, *entity.User) (*entity.User, error) {
	args := mock.Called()
	result := args.Get(1)
	return result.(*entity.User), args.Error(1)
}
func (mock *MockShorturlRepo) Delete(context.Context, *entity.User) error {
	args := mock.Called()
	//result := args.Get(0)
	return args.Error(1)
}

//func TestGet(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	mockRepo := new(MockShorturlRepo)
//	var del bool = false
//	var url string = "https://examp.ru"
//	var slug string = "1674872720465761244B_5"
//	var userId string = "1234"
//	sh := entity.Gofermart{Slug: slug, URL: url, UserID: userId, Del: del}
//	//Setup expectations
//	mockRepo.On("Get").Return(sh, nil)
//
//	testInFiles := NewInFiles(cfg)
//
//	result, _ := testInFiles.Get(ctx, &sh)
//
//	//Mock Assertion: Behavioral
//	mockRepo.AssertExpectations(t)
//
//	//Data Assertion
//	assert.Equal(t, slug, result.Slug)
//	assert.Equal(t, url, result.URL)
//	assert.Equal(t, userId, result.UserID)
//	assert.Equal(t, del, result.Del)
//}
