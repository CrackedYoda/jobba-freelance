export interface ApiEnvelope<T> { success: boolean; message?: string; data: T }
export interface Page<T> { docs?: T[]; data?: T[]; totalDocs?: number; totalCount?: number; page?: number; totalPages?: number }
export interface Job { _id: string; jobTitle: string; category: string; description: string; minBudget: number; maxBudget: number; paymentMethod: 'escrow'|'direct'; startDate: string; endDate: string; workType: 'onsite'|'remote'; location: string; applicantType: 'all'|'top-rated'|'invite'; privacy: 'public'|'private'; status: string; skills?: { _id: string; name: string }[]; createdAt: string }
export interface Worker { _id: string; firstName: string; lastName: string; profileImage?: string; country?: string; stateOfResidence?: string; lga?: string; workType?: string[]; skills?: { _id: string; name: string }[] }
export interface Application { _id: string; job: Job | Record<string, unknown>; worker: Worker | Record<string, unknown>; coverLetter: string; bidAmount: number; workDuration: string; status: string; createdAt: string }
export interface Payment { _id: string; reference: string; job: string; application: string; amount: number; currency: string; status: string; paymentMethod: string; createdAt: string }
export interface UserProfile { _id: string; firstName: string; lastName: string; email: string; phoneNumber: string; accountType: 'client'|'worker'; profileImage?: string; country?: string; stateOfResidence?: string; lga?: string; locationAddress?: string }
export interface CreateJob { jobTitle: string; category: string; description: string; skills: string[]; attachments: string[]; minBudget: number; maxBudget: number; paymentMethod: 'escrow'|'direct'; startDate: string; endDate: string; workType: 'onsite'|'remote'; location: string; applicantType: 'all'|'top-rated'|'invite'; privacy: 'public'|'private' }
export interface Notification { _id:string; title:string; message:string; type:string; isRead:boolean; createdAt:string }
export interface Dispute { _id:string; job:string; reason:string; description:string; evidenceUrls:string[]; status:string; resolutionMessage?:string; createdAt:string }
export interface ChatUser { _id:string; firstName:string; lastName:string; accountType?:string }
export interface Conversation { _id:string; participants:ChatUser[]; jobId:{_id:string;jobTitle:string;status:string}; lastMessage?:Message; lastMessageAt:string; isArchived:boolean; isUrgent:boolean }
export interface Message { _id:string; conversationId:string; senderId:ChatUser|string; receiverId:ChatUser|string; messageType:string; content?:string; isRead:boolean; createdAt:string }
export interface Offer { _id:string; jobId:string; clientId:string; workerId:string; budget:number; title?:string; description?:string; location?:string; status:string; createdAt:string }
export interface JobContext { jobId:string; status:string; role:string; counterparty?:{_id?:string;firstName?:string;lastName?:string}; canReview:boolean; hasReviewed:boolean; canDispute:boolean; hasOpenDispute:boolean }
export interface Review { _id:string; rating:number; comment:string; createdAt:string }
export interface ReviewResult { reviews:Review[]; totalReviews:number; averageRating:number }
