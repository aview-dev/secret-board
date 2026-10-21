import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	// The Google OAuth redirect URI is registered for :5173. Fail loudly when the
	// port is taken instead of silently moving to :5174 and breaking sign-in.
	// host: true listens on every interface (IPv4 + IPv6). Without it vite binds
	// only to what `localhost` resolves to (here [::1]), so SSH / VS Code port
	// forwarding to 127.0.0.1:5173 is refused.
	server: { host: true, port: 5173, strictPort: true },
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter()
		})
	]
});
