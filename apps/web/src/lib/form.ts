import { toastError } from '@laurel/ui/toast/index.js'
import { defaults, type FormOptions, type Infer, superForm } from 'sveltekit-superforms'
import { type ZodValidationSchema, zod4, zod4Client } from 'sveltekit-superforms/adapters'

type Options<S extends ZodValidationSchema> = Omit<
	Partial<FormOptions<Infer<S, 'zod4'>>>,
	'SPA' | 'validators' | 'onUpdate'
>

export function createForm<S extends ZodValidationSchema>(
	schema: S,
	onSubmit: (data: Infer<S, 'zod4'>) => Promise<unknown>,
	options?: Options<S>,
) {
	return superForm(defaults(zod4(schema)), {
		...options,
		SPA: true,
		validators: zod4Client(schema),
		async onUpdate({ form }) {
			if (!form.valid) return
			try {
				await onSubmit(form.data)
			} catch (error) {
				toastError(error)
			}
		},
	})
}
