import request from './request'
import axios from 'axios'

export const authApi = {
  checkInit() {
    return request.get('/auth/check-init') as Promise<{ need_init: boolean }>
  },

  init(username: string, password: string) {
    return request.post('/auth/init', { username, password }) as Promise<{ message: string; user: any }>
  },

  login(username: string, password: string) {
    return request.post('/auth/login', { username, password }) as Promise<{
      message: string
      access_token: string
      refresh_token: string
      user: any
    }>
  },

  logout() {
    return request.post('/auth/logout') as Promise<{ message: string }>
  },

  refresh() {
    const refreshToken = localStorage.getItem('refresh_token')
    return axios.post('/api/auth/refresh', null, {
      headers: { Authorization: `Bearer ${refreshToken}` }
    }).then(res => res.data) as Promise<{ access_token: string }>
  },

  getUser() {
    return request.get('/auth/user') as Promise<{ user: any }>
  },

  changePassword(oldPassword: string, newPassword: string) {
    return request.put('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword
    }) as Promise<{ message: string }>
  },

  captchaConfig() {
    return request.get('/auth/captcha-config') as Promise<{ enabled: boolean; captcha_id: string }>
  }
}
