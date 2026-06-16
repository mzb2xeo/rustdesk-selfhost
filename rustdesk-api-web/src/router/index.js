import { createRouter, createWebHashHistory, RouterView } from 'vue-router'

const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    meta: { title: 'Login' },
    component: () => import('@/views/login/login.vue'),
  },
  {
    path: '/register',
    name: 'Register',
    meta: { title: 'Register' },
    component: () => import('@/views/register/index.vue'),
  },
  {
    path: '/404',
    component: () => import('@/views/error-page/404.vue'),
    hidden: true,
  },
  {
    path: '/oauth/:code',
    meta: { title: 'OauthLogin' },
    component: () => import('@/views/oauth/login.vue'),
    hidden: true,
  },
  {
    path: '/oauth/bind/:code',
    meta: { title: 'OauthBind' },
    component: () => import('@/views/oauth/bind.vue'),
    hidden: true,
  },
]
export const asyncRoutes = [
  {
    path: '/my',
    name: 'My',
    redirect: '/',
    meta: { title: 'My', icon: 'UserFilled' },
    component: () => import('@/layout/index.vue'),
    children: [
      {
        path: '/',
        name: 'MyInfo',
        meta: { title: 'Userinfo', icon: 'User' },
        component: () => import('@/views/my/info.vue'),
      },
      {
        path: '/my/client_config',
        name: 'MyClientConfig',
        meta: { title: 'ClientConfig', icon: 'Cpu' },
        component: () => import('@/views/my/client_config.vue'),
      },
      {
        path: '/my/peer',
        name: 'MyPeer',
        meta: { title: 'MyPeer', icon: 'Monitor' },
        component: () => import('@/views/my/peer/index.vue'),
      },
      {
        path: '/my/address_book',
        name: 'MyAddressBookList',
        meta: { title: 'AddressBooks', icon: 'Notebook' },
        component: () => import('@/views/my/address_book/index.vue'),
      },
      {
        path: '/my/ab_settings',
        name: 'MyAbSettings',
        meta: { title: 'MyAbSettings', icon: 'Setting' },
        component: RouterView,
        children: [
          {
            path: '/my/address_book_collection',
            name: 'MyAddressBookCollection',
            meta: { title: 'AddressBookName', icon: 'Collection' },
            component: () => import('@/views/my/address_book/collection.vue'),
          },
          {
            path: '/my/tag',
            name: 'MyTagList',
            meta: { title: 'Tags', icon: 'CollectionTag' },
            component: () => import('@/views/my/tag/index.vue'),
          },
        ]
      },
      {
        path: '/my/logs',
        name: 'MyLogs',
        meta: { title: 'MyLogs', icon: 'List' },
        component: RouterView,
        children: [
          {
            path: '/my/shareRecord',
            name: 'MyShareRecordList',
            meta: { title: 'ShareRecord', icon: 'Share' },
            component: () => import('@/views/my/share_record/index.vue'),
          },
          {
            path: '/my/loginLog',
            name: 'MyLoginLog',
            meta: { title: 'LoginLog', icon: 'List' },
            component: () => import('@/views/my/login_log/index.vue'),
          },
        ]
      }
    ],
  },
  {
    path: '/user',
    name: 'User',
    redirect: '/user/index',
    meta: { title: 'System', icon: 'Setting' },
    component: () => import('@/layout/index.vue'),
    children: [
      {
        path: '/user/devices_connections',
        name: 'DevicesConnections',
        meta: { title: 'DevicesConnections', icon: 'Monitor' },
        component: RouterView,
        children: [
          {
            path: '/user/peer',
            name: 'Peer',
            meta: { title: 'PeerManage', icon: 'Monitor' },
            component: () => import('@/views/peer/index.vue'),
          },
          {
            path: '/user/deviceGroup',
            name: 'DeviceGroup',
            meta: { title: 'DeviceGroupManage', icon: 'ChatRound' },
            component: () => import('@/views/group/deviceGroupList.vue'),
          },
          {
            path: '/user/auditConn',
            name: 'AuditConn',
            meta: { title: 'AuditConnLog', icon: 'Tickets' },
            component: () => import('@/views/audit/connList.vue'),
          },
          {
            path: '/user/auditFile',
            name: 'AuditFile',
            meta: { title: 'AuditFileLog', icon: 'Files' },
            component: () => import('@/views/audit/fileList.vue'),
          },
          {
            path: '/user/serverCmd',
            name: 'ServerCmd',
            meta: { title: 'ServerCmd', icon: 'Tools' },
            component: () => import('@/views/rustdesk/control.vue'),
          },
        ]
      },
      {
        path: '/user/address_books',
        name: 'AddressBooksSystem',
        meta: { title: 'AddressBooksSystem', icon: 'Notebook' },
        component: RouterView,
        children: [
          {
            path: '/user/addressBook',
            name: 'UserAddressBook',
            meta: { title: 'AddressBookManage', icon: 'Notebook' },
            component: () => import('@/views/address_book/index.vue'),
          },
          {
            path: '/user/addressBookName',
            name: 'UserAddressBookName',
            meta: { title: 'AddressBookNameManage', icon: 'Collection' },
            component: () => import('@/views/address_book/collection.vue'),
          },
          {
            path: '/user/tag',
            name: 'UserTag',
            meta: { title: 'TagsManage', icon: 'CollectionTag' },
            component: () => import('@/views/tag/index.vue'),
          },
          {
            path: '/user/shareRecord',
            name: 'ShareRecord',
            meta: { title: 'ShareRecord', icon: 'Share' },
            component: () => import('@/views/share_record/index.vue'),
          },
        ]
      },
      {
        path: '/user/users_security',
        name: 'UsersSecurity',
        meta: { title: 'UsersSecurity', icon: 'User' },
        component: RouterView,
        children: [
          {
            path: '/user/index',
            name: 'UserList',
            meta: { title: 'UserManage', icon: 'User' },
            component: () => import('@/views/user/index.vue'),
          },
          {
            path: '/user/add',
            name: 'UserAdd',
            meta: { title: 'UserAdd', hide: true },
            component: () => import('@/views/user/edit.vue'),
          },
          {
            path: '/user/edit/:id',
            name: 'UserEdit',
            meta: { title: 'UserEdit', hide: true },
            component: () => import('@/views/user/edit.vue'),
          },
          {
            path: '/user/group',
            name: 'UserGroup',
            meta: { title: 'GroupManage', icon: 'ChatRound' },
            component: () => import('@/views/group/index.vue'),
          },
          {
            path: '/user/loginLog',
            name: 'LoginLog',
            meta: { title: 'LoginLog', icon: 'List' },
            component: () => import('@/views/login/log.vue'),
          },
          {
            path: '/user/userToken',
            name: 'UserToken',
            meta: { title: 'UserToken', icon: 'Ticket' },
            component: () => import('@/views/user/token.vue'),
          },
          {
            path: '/user/oauth',
            name: 'Oauth',
            meta: { title: 'OauthManage', icon: 'Link' },
            component: () => import('@/views/oauth/index.vue'),
          },
        ]
      },
    ],
  },
]
export const lastRoutes = [
  { path: '/:catchAll(.*)', redirect: '/404', meta: { hide: true } },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes: constantRoutes,
})


