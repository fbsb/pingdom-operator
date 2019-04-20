/*
Copyright 2019 Fabian Sabau <fabian.sabau@gmail.com>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpcheck

import (
	"errors"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

var (
	ErrAlreadyInitialized = errors.New("the httpcheck service has already been initialized")
	ErrNotInitialized     = errors.New("the httpcheck service has not been initialized")
)

type Service interface {
	Create(check pingdom.Check) (*pingdom.CheckResponse, error)
	Update(id int, check pingdom.Check) (*pingdom.PingdomResponse, error)
	Delete(id int) (*pingdom.PingdomResponse, error)
}

var instance Service

func InitService(client *pingdom.Client) error {
	if instance == nil {
		instance = client.Checks
		return nil
	}

	return ErrAlreadyInitialized
}

func ServiceInstance() (Service, error) {
	if instance != nil {
		return instance, nil
	}

	return nil, ErrNotInitialized
}
