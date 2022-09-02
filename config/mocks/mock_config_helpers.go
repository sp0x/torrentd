package mocks

import gomock "github.com/golang/mock/gomock"

func GetMockedConfig(ctrl *gomock.Controller) *MockConfig {
	config := NewMockConfig(ctrl)
	config.EXPECT().GetInt("port").Return(3333).Times(1)
	config.EXPECT().GetString("hostname").Return("").Times(1)
	config.EXPECT().GetBytes("api_key").Return(nil).Times(1)
	//config.EXPECT().GetBytes("workerCount").Return(1)
	config.EXPECT().GetBool("verbose").Return(true)
	config.EXPECT().GetInt("workerCount").Return(5555)
	config.EXPECT().Get("indexLoader").Return(nil)
	return config
}
