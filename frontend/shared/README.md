# Frontend shared boundary

This directory is reserved for business-neutral design tokens, primitive UI
components, and public TypeScript types that are genuinely shared by both
frontends. It must not contain routes, authentication state, API clients, or
Admin/Web feature modules.

The Web and Admin applications intentionally own their business code so one
bundle cannot accidentally import the other application's pages or endpoints.
