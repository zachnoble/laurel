import { createClient } from '@connectrpc/connect'
import { AuthService } from '@laurel/rpc/auth/v1/auth_pb.js'
import { createTransport } from '@laurel/rpc-runtime/transport.js'
import { API_URL } from '$app/env/public'

export const transport = createTransport(API_URL)

export const authClient = createClient(AuthService, transport)
