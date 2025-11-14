/*
 * Copyright ©  sixh sixh@apache.org
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
 */

class Menu {
  title: string;
  describe: string;
  icon: string;
  children?: Menu[];
  comp: string | any;
  active?: boolean;
  parentTitle?: string;
  constructor(
    title: string,
    describe: string,
    icon: string,
    active?: boolean,
    children?: Menu[],
    comp?: any,
    parentTitle?: string
  ) {
    this.title = title;
    this.describe = describe;
    this.icon = icon;
    this.children = children;
    this.comp = comp || null;
    this.active = active
    this.parentTitle = parentTitle;
  }
}

const menus: Menu[] = [
  new Menu(
    // 使用 i18n 键而非硬编码文案
    'menu.onlineServer.title',
    'menu.onlineServer.description',
    'brook-Diagram-',
    true,
    [],
    () => import('@/views/serverlist/Index.vue'),
  ),
  new Menu(
    'menu.tunnelConfiguration.title',
    'menu.tunnelConfiguration.description',
    'brook-technology_usb-cable',
    false,
    [],
    () => import('@/views/proxys/Configuration.vue'),
  ),
  new Menu(
    'menu.mySetting.title',
    'menu.mySetting.description',
    'brook-Gear-',
    false,
    [],
    () => import('@/views/mysetting/MySetting.vue'),
    'Setting'
  ),
]

export { Menu, menus };
