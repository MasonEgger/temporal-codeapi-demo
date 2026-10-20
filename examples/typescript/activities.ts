// ABOUTME: Activities for the temporal-typescript Codapi sandbox Hello World.
// Activities run in the Worker's normal Node.js environment, outside the Workflow sandbox.

export async function composeGreeting(name: string): Promise<string> {
  return `Hello, ${name}!`;
}
