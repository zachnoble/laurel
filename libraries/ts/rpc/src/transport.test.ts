import { createClient } from '@connectrpc/connect'
import { AuthService } from '@laurel/rpc/auth/v1/auth_pb.js'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { createTransport } from './transport.js'

describe('createTransport', () => {
	afterEach(() => vi.unstubAllGlobals())

	it('sends cookies, and uses GET for side-effect-free methods', async () => {
		const fetch = vi.fn(
			async () => new Response('{}', { headers: { 'Content-Type': 'application/json' } }),
		)
		vi.stubGlobal('fetch', fetch)

		await createClient(AuthService, createTransport('https://api.example.com')).getSession({})

		const [url, init] = fetch.mock.calls[0] as unknown as [string, RequestInit]
		expect(init.credentials).toBe('include')
		expect(init.method).toBe('GET')
		expect(url).toMatch(/^https:\/\/api\.example\.com\/auth\.v1\.AuthService\/GetSession\?/)
	})
})
