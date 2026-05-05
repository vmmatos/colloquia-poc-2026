import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Email inválido').max(254),
  password: z.string().min(8, 'Mínimo 8 caracteres').max(128),
});

export const registerSchema = loginSchema;

export const createChannelSchema = z.object({
  name: z.string().min(1, 'Campo obrigatório').max(80, 'Máximo 80 caracteres'),
  description: z.string().max(500, 'Máximo 500 caracteres').optional(),
  is_private: z.boolean(),
  type: z.enum(['channel', 'group']),
  initial_member_ids: z.array(z.string()).optional(),
});

export const messageSchema = z.object({
  content: z.string().min(1).max(4000, 'Máximo 4000 caracteres'),
});

export const profileSchema = z.object({
  name: z.string().max(100).optional(),
  bio: z.string().max(500).optional(),
  language: z.string().max(10).optional(),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type CreateChannelInput = z.infer<typeof createChannelSchema>;
export type ProfileInput = z.infer<typeof profileSchema>;
