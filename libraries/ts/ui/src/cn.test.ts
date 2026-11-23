import { describe, expect, it } from 'vitest'

import { cn } from './cn.js'

describe('cn', () => {
	it('merges conflicting tailwind classes, last wins', () => {
		expect(cn('px-2 py-1', { 'px-4': true, hidden: false })).toBe('py-1 px-4')
	})
})
