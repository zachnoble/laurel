import { defineEnvVars } from '@sveltejs/kit/env'
import { z } from 'zod'

export const variables = defineEnvVars({
	API_URL: {
		public: true,
		static: true,
		schema: z.url(),
		description: 'Base URL of the Go backend API',
	},
})
