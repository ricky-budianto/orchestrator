package dal

import (
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
)

func GetError(data *model.Error) error {
	return config.PostgreDB.First(data).Error
}
