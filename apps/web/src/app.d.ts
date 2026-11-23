import type { User } from '@laurel/rpc/auth/v1/auth_pb.js'

declare global {
	namespace App {
		interface PageData {
			user: User | null
		}
	}
}
