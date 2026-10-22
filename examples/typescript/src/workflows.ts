// ABOUTME: Workflow for the TypeScript IP geolocation tutorial, adapted from temporalio/temporal-tutorial-ipgeo-ts.
// Calls the two Activities in order with a Retry Policy, then builds the greeting.
// @@@SNIPSTART ts-ipgeo-workflow-imports
import * as workflow from '@temporalio/workflow';

// Only import the activity types
import type * as activities from './activities';

// Load Activities and assign the Retry Policy
const { getIP, getLocationInfo} = workflow.proxyActivities<typeof activities>({
  retry: {
    initialInterval: '1 second', // amount of time that must elapse before the first retry occurs.
    maximumInterval: '1 minute', // maximum interval between retries.
    backoffCoefficient: 2, // how much the retry interval increases.
    // maximumAttempts: 5, // maximum number of execution attempts. Unspecified means unlimited retries.
  },
  startToCloseTimeout: '1 minute', // maximum time allowed for a single Activity Task Execution.
});
// @@@SNIPEND

// @@@SNIPSTART ts-ipgeo-workflow-code
// The Temporal Workflow.
// Just a TypeScript function.
export async function getAddressFromIP(name: string): Promise<string> {

  try {
    const ip = await getIP();
    try {
      const location = await getLocationInfo(ip);
      return `Hello, ${name}. The machine running this code has the IP address ${ip} and is located in ${location}`;
    } catch (e) {
      throw new workflow.ApplicationFailure("Failed to get location");
    }
  } catch (e) {
    throw new workflow.ApplicationFailure("Failed to get IP");
  }

}
// @@@SNIPEND
