import { AxiosError } from 'axios';
import api from './api';
import type { ApiEnvelope, Application, Conversation, CreateJob, Dispute, Job, JobContext, Message, Notification, Offer, Page, Payment, Review, ReviewResult, UserProfile, Worker } from '../types/marketplace';

const message = (error: unknown) => error instanceof AxiosError
  ? ((error.response?.data as { message?: string } | undefined)?.message || (error.request ? 'Cannot reach Jobba. Check the API URL and connection.' : error.message))
  : error instanceof Error ? error.message : 'Something went wrong.';

async function request<T>(call: Promise<{ data: ApiEnvelope<T> }>): Promise<T> {
  try { const response = await call; return response.data.data; } catch (error) { throw new Error(message(error)); }
}
const items = <T>(page: Page<T>) => page.docs ?? page.data ?? [];

export const marketplace = {
  clientJobs: async () => items(await request<Page<Job>>(api.get('/jobs/client'))),
  workerJobs: async (search = '') => items(await request<Page<Job>>(api.get('/jobs/worker/search', { params: search ? { search } : undefined }))),
  savedJobs: async () => items(await request<Page<Job>>(api.get('/jobs/worker/saved'))),
  createJob: (body: CreateJob) => request<Job>(api.post('/jobs/client', body)),
  saveJob: (id: string) => request<unknown>(api.post(`/jobs/worker/${id}/save`)),
  unsaveJob: (id: string) => request<unknown>(api.delete(`/jobs/worker/${id}/save`)),
  apply: (id: string, body: { coverLetter: string; bidAmount: number; workDuration: string; attachments: string[] }) => request<Application>(api.post(`/jobs/worker/${id}/apply`, body)),
  applications: async () => items(await request<Page<Application>>(api.get('/jobs/worker/applications'))),
  workers: async (search = '', saved = false) => {
    const path = saved ? '/jobs/client/workers/saved-workers' : '/jobs/client/workers';
    const page = await request<Page<Worker>>(api.get(path, { params: search ? { search } : undefined }));
    return items(page);
  },
  toggleWorker: (id: string) => request<unknown>(api.post(`/jobs/client/workers/${id}/save`)),
  profile: () => request<UserProfile>(api.get('/auth/profile')),
  updateProfile: (body: Partial<UserProfile>) => request<UserProfile>(api.put('/auth/profile', body)),
  payments: async () => items(await request<Page<Payment>>(api.get('/payment'))),
  notifications: async () => items(await request<Page<Notification>>(api.get('/notifications'))),
  readNotification: (id:string) => request<Notification>(api.patch(`/notifications/${id}/read`)),
  readAllNotifications: () => request<unknown>(api.patch('/notifications/read-all')),
  disputes: async () => items(await request<Page<Dispute>>(api.get('/disputes'))),
  createDispute: (body:{jobId:string;reason:string;description:string;evidenceUrls:string[]}) => request<Dispute>(api.post('/disputes',body)),
  addDisputeEvidence: (id:string,evidenceUrls:string[]) => request<Dispute>(api.post(`/disputes/${id}/evidence`,{evidenceUrls})),
  conversations: async () => items(await request<Page<Conversation>>(api.get('/chat/conversations'))),
  messages: async (id:string) => items(await request<Page<Message>>(api.get(`/chat/conversations/${id}/messages`))),
  sendMessage: (id:string,content:string) => request<Message>(api.post(`/chat/conversations/${id}/messages`,{content,messageType:'text'})),
  markConversationRead: (id:string) => request<unknown>(api.post(`/chat/conversations/${id}/read`)),
  offers: async (role:'client'|'worker') => items(await request<Page<Offer>>(role === 'client' ? api.get('/jobs/client/offers') : api.get('/jobs/worker/offers'))),
  offer: (role:'client'|'worker',id:string) => request<Offer>(role === 'client' ? api.get(`/jobs/client/offers/${id}`) : api.get(`/jobs/worker/offers/${id}`)),
  acceptOffer: (id:string) => request<Offer>(api.post(`/jobs/worker/offers/${id}/accept`)),
  rejectOffer: (id:string) => request<unknown>(api.post(`/jobs/worker/offers/${id}/reject`)),
  jobContext: (id:string) => request<JobContext>(api.get(`/jobs/context/${id}`)),
  createReview: (role:'client'|'worker',jobId:string,rating:number,comment:string) => request<Review>(role === 'client' ? api.post(`/jobs/client/${jobId}/review`,{rating,comment}) : api.post(`/jobs/worker/${jobId}/review`,{rating,comment})),
  reviews: (role:'client'|'worker',userId:string) => request<ReviewResult>(role === 'client' ? api.get(`/jobs/client/reviews/${userId}`) : api.get(`/jobs/worker/reviews/${userId}`)),
};
