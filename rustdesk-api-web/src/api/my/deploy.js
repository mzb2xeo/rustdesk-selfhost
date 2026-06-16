import request from '@/utils/request'

export function createDeployToken () {
  return request({
    url: '/my/deploy/token',
    method: 'post',
  })
}
