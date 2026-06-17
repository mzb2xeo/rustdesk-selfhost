import request from '@/utils/request'

export function createDeployToken (data = {}) {
  return request({
    url: '/my/deploy/token',
    method: 'post',
    data,
  })
}

export function listDeployTokens (params = {}) {
  return request({
    url: '/my/deploy/token/list',
    method: 'get',
    params,
  })
}

export function revokeDeployToken (data = {}) {
  return request({
    url: '/my/deploy/token/revoke',
    method: 'post',
    data,
  })
}
