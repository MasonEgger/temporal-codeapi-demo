// ABOUTME: Workflows for the temporal-typescript Codapi sandbox Hello World.
// The Worker bundles this file into its deterministic Workflow sandbox at startup.
import { proxyActivities } from '@temporalio/workflow';
import type * as activities from './activities.ts';

const { composeGreeting } = proxyActivities<typeof activities>({
  startToCloseTimeout: '10 seconds',
});

export async function helloWorkflow(name: string): Promise<string> {
  return await composeGreeting(name);
}
