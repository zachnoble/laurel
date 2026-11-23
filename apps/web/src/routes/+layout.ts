import { createQueryOptions } from '@connectrpc/connect-query-core'
import { AuthService } from '@laurel/rpc/auth/v1/auth_pb.js'

import { queryClient } from '#lib/query-client.js'
import { transport } from '#lib/rpc-clients.js'
import type { LayoutLoad } from './$types'

export const ssr = false

export const load: LayoutLoad = async () => {
	const session = await queryClient
		.query(createQueryOptions(AuthService.method.getSession, {}, { transport }))
		.catch(() => null)
	return { user: session?.user ?? null }
}
