import axios from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

export const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle auth errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  }
)

// Auth API
export const authApi = {
  register: async (data: {
    email: string
    password: string
    first_name: string
    last_name: string
    org_name: string
    org_slug: string
  }) => {
    const response = await apiClient.post('/auth/register', data)
    return response.data
  },

  login: async (data: { email: string; password: string }) => {
    const response = await apiClient.post('/auth/login', data)
    return response.data
  },

  logout: async () => {
    await apiClient.post('/auth/logout')
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  },

  me: async () => {
    const response = await apiClient.get('/auth/me')
    return response.data
  },
}

// Mailbox API
export const mailboxApi = {
  list: async () => {
    const response = await apiClient.get('/mailboxes')
    return response.data
  },

  create: async (data: {
    domain_id: string
    local_part: string
    password: string
    display_name: string
    quota_bytes: number
  }) => {
    const response = await apiClient.post('/mailboxes', data)
    return response.data
  },

  get: async (id: string) => {
    const response = await apiClient.get(`/mailboxes/${id}`)
    return response.data
  },

  update: async (id: string, data: Partial<{
    display_name: string
    password: string
    quota_bytes: number
  }>) => {
    const response = await apiClient.patch(`/mailboxes/${id}`, data)
    return response.data
  },

  delete: async (id: string) => {
    await apiClient.delete(`/mailboxes/${id}`)
  },

  suspend: async (id: string) => {
    const response = await apiClient.post(`/mailboxes/${id}/suspend`)
    return response.data
  },

  unsuspend: async (id: string) => {
    const response = await apiClient.post(`/mailboxes/${id}/unsuspend`)
    return response.data
  },
}

// Domain API
export const domainApi = {
  list: async () => {
    const response = await apiClient.get('/domains')
    return response.data
  },

  create: async (data: { name: string }) => {
    const response = await apiClient.post('/domains', data)
    return response.data
  },

  get: async (id: string) => {
    const response = await apiClient.get(`/domains/${id}`)
    return response.data
  },

  delete: async (id: string) => {
    await apiClient.delete(`/domains/${id}`)
  },

  verify: async (id: string) => {
    const response = await apiClient.post(`/domains/${id}/verify`)
    return response.data
  },

  regenerate: async (id: string) => {
    const response = await apiClient.post(`/domains/${id}/regenerate`)
    return response.data
  },
}

// Organization API
export const organizationApi = {
  list: async () => {
    const response = await apiClient.get('/organizations')
    return response.data
  },

  get: async (id: string) => {
    const response = await apiClient.get(`/organizations/${id}`)
    return response.data
  },

  update: async (id: string, data: { name: string; slug: string }) => {
    const response = await apiClient.patch(`/organizations/${id}`, data)
    return response.data
  },
}
