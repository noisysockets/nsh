/*
 * Copyright 2024 Damian Peckett <damian@pecke.tt>
 *
 * Licensed under the Noisy Sockets Source License 1.0 (NSSL-1.0); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at
 *
 * https://github.com/noisysockets/nsh/blob/main/LICENSE
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var webFS embed.FS

func Handler() http.Handler {
	subFS, err := fs.Sub(webFS, "dist")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(subFS))
}
