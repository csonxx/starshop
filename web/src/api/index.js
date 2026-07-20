import { http } from './http'

export const api = {
  sendCode: (phone) => http.post('/auth/send-code', { phone }),
  login: (phone, code) => http.post('/auth/login', { phone, code }),
  me: () => http.get('/me'),

  banners: () => http.get('/banners'),
  tags: (type) => http.get('/tags', { params: type ? { type } : {} }),
  cases: (params) => http.get('/cases', { params }),
  caseDetail: (id) => http.get(`/cases/${id}`),
  pinned: () => http.get('/cases/pinned'),

  // admin
  adminOverview: () => http.get('/admin/overview'),
  adminStatsByStyle: () => http.get('/admin/stats/by-style'),

  adminListBanners: () => http.get('/admin/banners'),
  adminCreateBanner: (payload) => http.post('/admin/banners', payload),
  adminUpdateBanner: (id, payload) => http.put(`/admin/banners/${id}`, payload),
  adminDeleteBanner: (id) => http.delete(`/admin/banners/${id}`),

  adminListTags: (type) => http.get('/admin/tags', { params: type ? { type } : {} }),
  adminCreateTag: (payload) => http.post('/admin/tags', payload),
  adminUpdateTag: (id, payload) => http.put(`/admin/tags/${id}`, payload),
  adminDeleteTag: (id) => http.delete(`/admin/tags/${id}`),

  adminListCases: () => http.get('/admin/cases'),
  adminCreateCase: (payload) => http.post('/admin/cases', payload),
  adminUpdateCase: (id, payload) => http.put(`/admin/cases/${id}`, payload),
  adminDeleteCase: (id) => http.delete(`/admin/cases/${id}`),
  adminTogglePin: (id) => http.post(`/admin/cases/${id}/pin`)
}