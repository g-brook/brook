#!/usr/bin/env bash

#
# Copyright ©  sixh sixh@apache.org
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#!/usr/bin/env bash

set -e

echo "Building portal/server..."
rm -rf dist
cd ../portal/server
pnpm run build
echo "Copying dist to scmd/web ..."
rm -rf ../../scmd/web/dist/assets
rm -rf ../../scmd/web/dist/index.html
rm -rf ../../scmd/web/dist/logo.svg
cp -r dist ../../scmd/web/
cd ../../scmd
echo "Building Go binary..."

bash ./build-go.sh

echo "Done."