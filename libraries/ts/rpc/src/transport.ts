import type { Transport } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

/**
 * Connect transport for a Go service built on libraries/go/rpc: sends the session
 * cookie cross-origin (the Go CORS middleware allows credentials) and uses GET for
 * `idempotency_level = NO_SIDE_EFFECTS` methods so they can be cached.
 */
export function createTransport(baseUrl: string): Transport {
	return createConnectTransport({
		baseUrl,
		useHttpGet: true,
		fetch: (input, init) => fetch(input, { ...init, credentials: 'include' }),
	})
}
