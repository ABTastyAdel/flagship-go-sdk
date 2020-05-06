/****************************************************************************
 * Copyright 2020, Flagship, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package flagship

import (
	"github.com/abtasty/flagship-go-sdk/pkg/client"
)

// Start returns a FlagshipClient instantiated with the given envID and options
func Start(envID string, clientOptions ...client.OptionFunc) (*client.FlagshipClient, error) {
	factory := &client.FlagshipFactory{
		EnvID: envID,
	}
	return factory.CreateClient(clientOptions...)
}
