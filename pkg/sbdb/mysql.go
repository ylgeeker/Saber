/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package sbdb

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQL struct {
	db *gorm.DB
}

// DB returns the underlying *gorm.DB for advanced use.
func (m *MySQL) DB() *gorm.DB {
	return m.db
}

func NewMySQL(opts ...Option) (*MySQL, error) {
	o := defaultOptions

	for _, opt := range opts {
		if err := opt.apply(&o); err != nil {
			return nil, err
		}
	}

	db, err := gorm.Open(mysql.Open(o.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	return &MySQL{db: db}, nil
}
