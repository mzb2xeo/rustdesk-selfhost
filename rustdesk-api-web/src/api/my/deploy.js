import request from '@/utils/request'

export function createDeployToken (data = {}) {
  return request({
    url: '/my/deploy/token',
    method: 'post',
    data,
  })
}
