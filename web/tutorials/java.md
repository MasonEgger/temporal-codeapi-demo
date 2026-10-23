---
id: hello-world-tutorial-java
title: Build a Temporal Application from scratch in Java
sidebar_position: 3
description: In this tutorial you will build a Temporal Application using the Java SDK. You'll write a Workflow, an Activity, tests, and define a Worker.
keywords: [java,temporal,sdk,tutorial, hello world, temporal application, workflow, activity]
last_update:
  author: Brian P. Hogan
  date: 2024-12-17
tags:
  - java
  - sdk
  - tutorial
image: /img/temporal-logo-twitter-card.png
---

<img className="banner" src="/img/sdk_banners/banner_java.png" alt="Temporal Java SDK" />

:::note Tutorial information

- **Level:** ⭐ Temporal beginner
- **Time:** ⏱️ ~15 minutes
- **Goals:** 🙌
  - Set up, build, and test a Temporal Application project from scratch using the [Java SDK](https://github.com/temporalio/sdk-java).
  - Identify the four parts of a Temporal Workflow application.
  - Describe how the Temporal Server gets information to the Worker.
  - Observe how Temporal recovers from failed Activities.

:::

### Introduction

Creating reliable applications is a difficult chore. [Temporal](https://temporal.io/) lets you create fault-tolerant, resilient applications using programming languages you already know, so you can build complex applications that execute reliably and recover from failures.

In this tutorial, you will build your first Temporal Application from scratch using the [Temporal Java SDK](https://github.com/temporalio/sdk-java). You'll develop a small application that asks for your name and then uses APIs to get the public IP address of the machine running the code and the location based on that address. External requests can fail due to rate limiting, network interruptions, or other errors. Using Temporal for this application will let you automatically recover from these and other kinds of failures without having to write explicit error-handling code.

The app will consist of the following pieces:

1. Two [Activities](https://docs.temporal.io/activities): Activities are functions called by Workflows, and they contain any logic that might fail or behave differently at different times. The first Activity will get your IP address, and the second Activity will use that IP address to get your location.

1. A [Workflow](https://docs.temporal.io/workflows): Workflows are functions that define the overall flow of the application.  Your Workflow will execute both Activities you define. It will call the first Activity to fetch the IP address, and then use the result of that Activity to call the next Activity to find your location.
2. A [Worker](https://docs.temporal.io/workers): Workers host the Activity and Workflow code and execute the code piece by piece.
3. A client program: You'll need to start your Workflow. In this tutorial you'll create a small program using the [Temporal Client](https://docs.temporal.io/encyclopedia/temporal-sdks#temporal-client) to start the Workflow.

You'll also write tests to ensure your Workflow executes successfully.

When you're done, you'll have a basic application and a clear understanding of how to build out the components you'll need in future Temporal applications in Java.

## Prerequisites

Before starting this tutorial:

- Follow the tutorial [Run your first Temporal application with the Java SDK](https://learn.temporal.io/getting_started/java/first_program_in_java/) to gain a better understanding of what Temporal is and how its components fit together.

:::note Playground
This playground runs a Temporal Service, the Temporal Web UI, and your Worker for you, so you don't need to install anything. When you're ready to build on your own machine, [set up a local development environment for developing Temporal applications using Java](https://learn.temporal.io/getting_started/java/dev_environment/).
:::

## Create a New Temporal Java Project

To get started with the Temporal Java SDK, you'll create a new Maven project, just like any other Java program you're creating. Then you'll add the Temporal SDK package to your project.

The playground has already done these steps for you. Open them when you're ready to create the project on your own machine.

:::setup Create the project on your own machine
In a terminal, create a new project directory called `temporal-ip-geolocation`:

```command
mkdir temporal-ip-geolocation
```

Switch to the new directory:

```command
cd temporal-ip-geolocation
```

From the root of your new project directory, create the directories for your application code and your tests, using the standard Maven layout and an `iplocate` package:

```command
mkdir -p src/main/java/iplocate src/test/java/iplocate
```

Then create a `pom.xml` file in the project root that adds the Temporal Java SDK as a project dependency, along with the testing libraries you'll use later in this tutorial:

<!--SNIPSTART java-ipgeo-pom-->
[pom.xml](pom.xml)
<!--SNIPEND-->

The Surefire configuration loads Mockito as a Java agent when your tests run, which Mockito needs on recent Java versions to create mocks.
:::

With the project created, you'll create the application's core logic.

## Write functions to call external services

Your application will make two HTTP requests. The first request will return your current public IP, while the second request will use that IP to provide city, state, and country information.

You'll use Temporal Activities to make these requests. You use Activities in your Temporal Applications to execute [non-deterministic](https://docs.temporal.io/workflows#deterministic-constraints) code or perform operations that may fail, such as API requests or database calls.

If an Activity fails, Temporal can automatically retry it until it succeeds or reaches a specified retry limit. This ensures that transient issues, like network glitches or temporary service outages, don't result in data loss or incomplete processes.

With the Temporal Java SDK, you define Activities with an interface and an implementation. Create the file `src/main/java/iplocate/IPActivities.java`:

```command
touch src/main/java/iplocate/IPActivities.java
```

Add the following code to define the Activity interface:

<!--SNIPSTART java-ipgeo-activity-interface-->
[src/main/java/iplocate/IPActivities.java](src/main/java/iplocate/IPActivities.java)
<!--SNIPEND-->

The `@ActivityInterface` annotation tells Temporal that the methods in this interface are Activities. Your Workflow will call these methods, and your Worker will run the class that implements them.

Now create the file `src/main/java/iplocate/IPActivitiesImpl.java`:

```command
touch src/main/java/iplocate/IPActivitiesImpl.java
```

Open the file `src/main/java/iplocate/IPActivitiesImpl.java` in your editor and add the following code that imports dependencies and defines a class to hold the Activities:


<!--SNIPSTART java-ipgeo-activity-setup-->
[src/main/java/iplocate/IPActivitiesImpl.java](src/main/java/iplocate/IPActivitiesImpl.java)
<!--SNIPEND-->

With the Java SDK, you implement Activities as methods on a class. Because it's a class, you can pass shared objects like database connections through its constructor. In this tutorial you'll pass an HTTP client into the constructor. This will let you stub out the HTTP client implementation when you write tests for your Activities later.

Now add the following code to define a Temporal Activity that retrieves your IP address from `icanhazip.com`:

<!--SNIPSTART java-ipgeo-activity-ip-->
[src/main/java/iplocate/IPActivitiesImpl.java](src/main/java/iplocate/IPActivitiesImpl.java)
<!--SNIPEND-->

Java's `HttpClient` throws checked exceptions, so the method catches them and rethrows them with `Activity.wrap`. This converts a checked exception into an unchecked one that Temporal reports as an Activity failure, without adding another layer to the error.

The response from `icanhazip.com` is plain-text, and it includes a newline, so you trim off the newline character before returning the result.

Notice that this method doesn't try to recover from errors. When you build your Workflow, you'll use Temporal's Activity Retry policies to retry this code automatically if there's an error.

Now add the second Activity that accepts an IP address and retrieves location data. In `src/main/java/iplocate/IPActivitiesImpl.java`, add the following code to define it:

<!--SNIPSTART java-ipgeo-activity-location-->
[src/main/java/iplocate/IPActivitiesImpl.java](src/main/java/iplocate/IPActivitiesImpl.java)
<!--SNIPEND-->

This Activity follows the same pattern as the `getIP` Activity.  This time, the service returns JSON data rather than text, so you use Jackson's `ObjectMapper`, which the Temporal Java SDK already depends on, to read the data.

While Activities can accept input arguments, it's a best practice to send a single argument rather than multiple arguments. In this case you only have a single String. If you have more than one argument, you should bundle them up in a serializable object. This is because later revisions to your Workflows and Activities that change the number of arguments sent to them can otherwise introduce the need for versioning. Review the [Activity parameters](https://docs.temporal.io/develop/java/activities/basics#activity-parameters) section of the Temporal documentation for more details.

You've created your two Activities. Now you'll coordinate them using a Temporal Workflow.

## Control application logic with a Workflow

Workflows are where you configure and organize the execution of Activities. You define a Workflow by writing a *Workflow Definition* using one of the Temporal SDKs.

In the Temporal Java SDK, a Workflow Definition is made of two parts: an interface annotated with `@WorkflowInterface` that contains one method annotated with `@WorkflowMethod`, and a class that implements that interface. The method's return value is the result of the Workflow Execution. Review the [Develop Workflows](https://docs.temporal.io/develop/java/workflows/basics#develop-workflows) section of the Temporal documentation for more about Workflows in Java.

Temporal Workflows [must be deterministic](https://docs.temporal.io/workflow-definition#deterministic-constraints) so that Temporal can replay your Workflow in the event of a crash. That's why you call Activities from your Workflow code. Activities don't have the same determinism constraints that Workflows have.

Create the file `src/main/java/iplocate/GetAddressFromIPWorkflow.java`:

```command
touch src/main/java/iplocate/GetAddressFromIPWorkflow.java
```

Add the following code to define the Workflow interface:

<!--SNIPSTART java-ipgeo-workflow-interface-->
[src/main/java/iplocate/GetAddressFromIPWorkflow.java](src/main/java/iplocate/GetAddressFromIPWorkflow.java)
<!--SNIPEND-->

The `@WorkflowMethod` annotation marks `getAddressFromIP` as the method that runs when the Workflow starts. The Workflow Execution completes when this method returns.

Now create the file `src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java`:

```command
touch src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java
```

Then add the following code to import the Activities and configure how the Workflow should handle failures with a [Retry Policy](https://docs.temporal.io/encyclopedia/retry-policies).

<!--SNIPSTART java-ipgeo-workflow-imports-->
[src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java](src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java)
<!--SNIPEND-->


With the imports and options in place, you can define the Workflow itself.

Add the following code to define the `GetAddressFromIPWorkflowImpl` class, which will call both Activities, using the value of the first as the input to the second:

<!--SNIPSTART java-ipgeo-workflow-code-->
[src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java](src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java)
<!--SNIPEND-->

`Workflow.newActivityStub` creates a stub for your Activity interface with the options you set. When the Workflow calls a method on the stub, Temporal schedules that Activity and returns its result.

The `getAddressFromIP` method accepts a string value that holds the name and returns a string value.  Like with Activities, you can send multiple inputs into a Workflow, but it's a good practice to combine those into a single object.

In this example, you have specified that the Start-to-Close Timeout for your Activities will be one minute, meaning that your Activity has one minute to complete before it times out. Of all the Temporal timeout options, `setStartToCloseTimeout` is the one you should always set.

You also set the Retry Policy for Activities. In this example, you're using the default Retry Policy values, so you don't need to specify the values, but by leaving them in you have a more clear picture of what happens. Note that the `setMaximumAttempts` is commented out, which means there's no limit to the number of times Temporal will retry your Activities if they fail.


This code does check for and handle errors, but because you've configured unlimited retries, there won't be any exceptions caught. However, if you change the Retry Policy's maximum retries, or you specify non-retryable exceptions, this code will be in place to handle those errors.

Next you'll create a Worker that executes the Workflow and Activities.

## Configure and run a Worker

When you start a Temporal Workflow, the Workflow and its Activities get scheduled on the Temporal Service's [Task Queue](https://docs.temporal.io/task-queue#task-queue). A [Worker](https://docs.temporal.io/workers) hosts Workflow and Activity functions and polls the Task Queue for tasks related to those Workflows and Activities. After the Worker runs the code, it communicates the results back to the Temporal Service where they're stored in the Event History. This records the Workflow's entire execution, enabling features like fault tolerance by allowing the Workflow to replay in case of Worker crashes or restarts.

You use the Temporal SDK to define a Worker Program.

In your Worker Program, you need to specify the name of the Task Queue, which must match the Task Queue name used whenever you interact with a Workflow from a client application. This ensures that the Worker processes tasks for the correct Workflow. The Task Queue name is a case-insensitive string. To ensure consistency and avoid errors, define the Task Queue name as a constant that can be reused throughout your code.

Create the file `src/main/java/iplocate/Shared.java`:

```command
touch src/main/java/iplocate/Shared.java
```

Open the file and add the following lines to the file to define the constant for the Task Queue:

<!--SNIPSTART java-ipgeo-shared-->
[src/main/java/iplocate/Shared.java](src/main/java/iplocate/Shared.java)
<!--SNIPEND-->

Now you can create the Worker program.

Create the file `src/main/java/iplocate/IPLocateWorker.java`:

```command
touch src/main/java/iplocate/IPLocateWorker.java
```

Then open `src/main/java/iplocate/IPLocateWorker.java` in your editor and add the following code to define the Worker program:

<!--SNIPSTART java-ipgeo-worker-->
[src/main/java/iplocate/IPLocateWorker.java](src/main/java/iplocate/IPLocateWorker.java)
<!--SNIPEND-->

The code lives in the `iplocate` package with your Workflow and Activity Definitions. It defines a `main` method that creates and runs a Worker.

You first create a client, and then you create a Worker that uses the client, along with the Task Queue it should listen on. `WorkflowServiceStubs.newLocalServiceStubs()` connects to the Temporal Cluster running at `localhost` on port `7233`, and the client uses the `default` namespace. You can change this by passing options to the service stubs and the client.

Then you register your Workflow and Activities with the Worker. You register the Workflow implementation class and an instance of your Activities class. This is also where you inject the HTTP client so your Activities can access it.

In this case your Worker will run your Workflow and your two Activities, but there are cases where you could configure one Worker to run Activities, and another Worker to run the Workflows.

Now you'll start the Worker. Be sure you have started the local Temporal Service and execute the following command to start your Worker:

```command
mvn compile exec:java \
    -Dexec.mainClass="iplocate.IPLocateWorker" \
    -Dorg.slf4j.simpleLogger.defaultLogLevel=warn
```

The Worker runs without printing any output.

In the playground, you don't start the Worker yourself. It starts in the background every time you run the client, which works the same as leaving it running in its own terminal.

Your Worker is running and is polling the Temporal Service for Workflows to run, but before you start your Workflow, you'll write tests to prove it works as expected.

## Write tests to ensure things work

The Temporal Java SDK includes functions that help you test your Workflow executions. Let's add a basic unit test to the application to make sure the Workflow works as expected.

You'll use [JUnit 5](https://junit.org/junit5/) to build your test cases and [Mockito](https://site.mockito.org/) to mock the Activity so you can test the Workflow in isolation. You added both, along with the `temporal-testing` package, to your `pom.xml` when you created the project.

Now create a file to hold the test for your Workflow. Create the file `src/test/java/iplocate/GetAddressFromIPWorkflowTest.java`:

```command
touch src/test/java/iplocate/GetAddressFromIPWorkflowTest.java
```

Add the following code to set up the testing environment:

<!--SNIPSTART java-ipgeo-workflow-test-setup-->
[src/test/java/iplocate/GetAddressFromIPWorkflowTest.java](src/test/java/iplocate/GetAddressFromIPWorkflowTest.java)
<!--SNIPEND-->

Add the following code to test the Workflow execution:

<!--SNIPSTART java-ipgeo-workflow-test-workflow-->
[src/test/java/iplocate/GetAddressFromIPWorkflowTest.java](src/test/java/iplocate/GetAddressFromIPWorkflowTest.java)
<!--SNIPEND-->

This test creates a test execution environment to run the Workflow.

Instead of using your actual Activities, you replace the Activities `getIP` and `getLocationInfo` with mocks that return hard-coded values. This way you're testing the Workflow's logic independently of the Activities. If you wanted to test the Activities directly as part of an integration test, you would omit these mocks from the test.  The `TestWorkflowExtension` passes the test a test environment, a Worker, and a Workflow stub, and you register the mock with the Worker before you start the environment.

The test then executes the Workflow in the test environment and checks for a successful execution. Finally, the test ensures the Workflow's return value returns the expected value.

Ensure you've saved all your files and execute your tests with the following command:

```command
mvn test
```

The test environment starts, spins up a Worker, and executes the Workflow in the test environment. At the end, you'll see that your test passes:

```output
[INFO] Running iplocate.GetAddressFromIPWorkflowTest
[INFO] Tests run: 1, Failures: 0, Errors: 0, Skipped: 0, Time elapsed: 1.571 s -- in iplocate.GetAddressFromIPWorkflowTest
[INFO] BUILD SUCCESS
```

In the playground, click **Run the Workflow test** to run the one test you have so far:

<div class="run-block" data-command="test" data-label="Run the Workflow test" data-output-label="JUnit output" data-omit-files="src/test/java/iplocate/IPActivitiesTest.java"></div>

With a Workflow test in place, you can write unit tests for the Activities.

Both of your Activities make external calls to services that will change their results based on who runs them. It will be challenging to test these Activities reliably. For example, the IP address may vary based on your machine's location.

To ensure you can test the Activities in isolation, you’ll stub out the HTTP calls.

Create the file `src/test/java/iplocate/IPActivitiesTest.java`

```command
touch src/test/java/iplocate/IPActivitiesTest.java
```

Add the following code to import the testing libraries and Activities you'll use, and then define a helper that builds a mock HTTP client and mock response:

<!--SNIPSTART java-ipgeo-activity-test-setup-->
[src/test/java/iplocate/IPActivitiesTest.java](src/test/java/iplocate/IPActivitiesTest.java)
<!--SNIPEND-->


To ensure you don't make real HTTP requests, the helper uses Mockito to create an `HttpClient` whose `send` method returns a mock response. When you write your tests you'll inject this mocked client into the Activity so you can control the response.

Next, write the test for the `getIP` Activity, using your mocked HTTP client to stub out actual HTTP calls so your tests are consistent. Notice that the mock response adds the newline character so it replicates the actual response:

<!--SNIPSTART java-ipgeo-activity-test-ip-->
[src/test/java/iplocate/IPActivitiesTest.java](src/test/java/iplocate/IPActivitiesTest.java)
<!--SNIPEND-->

Like in your Worker, you inject the HTTP client into the Activities class and then register the Activities with the test environment.

To test the Activity itself, you use the test environment to execute the Activity through a stub rather than directly calling the `getIP` method. You get the result from the Activity Execution and then ensure it matches the value you expect.

To test the `getLocationInfo` Activity, you use a similar approach. Add the following code to the `src/test/java/iplocate/IPActivitiesTest.java` file:

<!--SNIPSTART java-ipgeo-activity-test-location-->
[src/test/java/iplocate/IPActivitiesTest.java](src/test/java/iplocate/IPActivitiesTest.java)
<!--SNIPEND-->

This test looks similar to the previous test; you mock out the HTTP client and ensure it returns the expected data, and then you execute the Actiity in the test environment. Then you retrieve the value and ensure it's what you expect.

Run the tests again to see them pass. Maven lists each test class it runs, so you can see that all the tests ran:

```command
mvn test
```

```output
[INFO] Running iplocate.IPActivitiesTest
[INFO] Tests run: 2, Failures: 0, Errors: 0, Skipped: 0, Time elapsed: 3.931 s -- in iplocate.IPActivitiesTest
[INFO] Running iplocate.GetAddressFromIPWorkflowTest
[INFO] Tests run: 1, Failures: 0, Errors: 0, Skipped: 0, Time elapsed: 1.571 s -- in iplocate.GetAddressFromIPWorkflowTest
[INFO] Tests run: 3, Failures: 0, Errors: 0, Skipped: 0
[INFO] BUILD SUCCESS
```

In the playground, click **Run tests** to run all three tests:

<div class="run-block" data-command="test" data-label="Run tests" data-output-label="JUnit output"></div>

To see what a failing test looks like, this button changes the end of the expected greeting from `Planet Earth` to `Planet Mars` for one run:

<div class="run-block" data-command="test" data-label="Run tests with a wrong expected greeting" data-output-label="JUnit output" data-variant-file="src/test/java/iplocate/GetAddressFromIPWorkflowTest.java" data-variant-find="located in Planet Earth&quot;" data-variant-replace="located in Planet Mars&quot;"></div>

Now that you have your tests passing, it's time to start a Workflow Execution.

## Run the Workflow from a client

You can start a Workflow Execution by using the Temporal CLI or by writing code using the Temporal SDK.

Starting a Workflow Execution using the Temporal SDK involves connecting to the Temporal Server, configuring the Task Queue the Workflow should use, and starting the Workflow with the input parameters it expects. In a real application, you may invoke this code when someone submits a form, presses a button, or visits a certain URL. In this tutorial, you will create a small CLI program that runs your Temporal Workflow.

Create the file `src/main/java/iplocate/IPLocateClient.java`:

```command
touch src/main/java/iplocate/IPLocateClient.java
```

Open `src/main/java/iplocate/IPLocateClient.java` in your editor and add the following code to the file to connect to the server and start the Workflow:

<!--SNIPSTART java-ipgeo-cli-client-->
[src/main/java/iplocate/IPLocateClient.java](src/main/java/iplocate/IPLocateClient.java)
<!--SNIPEND-->

In the `main` method you check to see if there is at least one argument passed and then capture the user's name from the arguments.

The `main` method then sets up a connection to your Temporal Server, invokes your Workflow, passes in an argument for the `name`, and assigns the Workflow a unique identifier using a UUID. The client dispatches the Workflow on the same Task Queue that the Worker is polling on. That's why you used a constant to ensure the Task Queue name is consistent. If there's a mismatch, your Workflow will execute on a different Task Queue and there won't be any Workers polling for tasks.

:::note Specify a Workflow Id

A Workflow ID is unique in a Namespace and identifies a Workflow Execution. Using an identifier that reflects some business process or entity is a good practice. For example, you might use a customer identifier as part of the Workflow Id if you run one Workflow per customer. This would make it easier to find all Workflow Executions related to that customer later.

In this tutorial you're generating a UUID and appending it to a string that identifies the Workflow.
:::

You can [get the results](https://docs.temporal.io/develop/java/client/temporal-client#get-workflow-results) from your Workflow right away, or you can get the results at a later time. This implementation attempts to get the results immediately by calling `getAddressFromIP` on the Workflow stub, which blocks the program's execution until the Workflow Execution completes.

Now you can run your Workflow. First, ensure that your local Temporal Service is running, and that your Worker program is running also.

Then open a new terminal and switch to the project directory:

```command
cd temporal-ip-geolocation
```

Now run the following command to run the Workflow using the client program you wrote:

```command
mvn compile exec:java \
    -Dexec.mainClass="iplocate.IPLocateClient" \
    -Dexec.args="Ziggy" \
    -Dorg.slf4j.simpleLogger.defaultLogLevel=warn
```

You'll see the following output:

```output
Hello, Ziggy. The machine running this code has the IP address 203.0.113.7 and is located in Ashburn, Virginia, United States
```

In the playground, click **Run** to start your Worker and run the client with the name in the **Name** box:

<div class="run-block" data-command="run" data-history data-name data-output-label="Client output"></div>

:::note
In the playground, your Activities run on the server that hosts this page, so the IP address and location in the greeting belong to that server. When you run the application on your own machine, you'll see your own address and location.
:::

:::tip

To run your Temporal Application, you need to start the Workflow and the Worker. You can start these in any order, but you'll need to run each command from a separate terminal window, as the Worker needs to be constantly running to look for tasks to execute.

:::

Your Temporal Application works. Now review it in the Temporal Web UI.

## Exploring your application in the Web UI

The Temporal Web UI gives you insights into your Workflow's execution. Open the Temporal Web UI by visiting `http://localhost:8233` and click on your completed Workflow to view the execution history. In the playground, use the panel below, which shows the Temporal Service from your most recent run. You'll see results similar to the following image:

<div class="webui"></div>

![The UI shows the results of the Workflow including the output](images/overview.png)

You'll see the dates the Workflow Exeuction ran, how long it took to run, the input to the Workflow, and the result.

After that, you see the Event History, detailing the entire flow, including the inputs and outputs of the Activity Executions:

![The Workflow History showing all Activities and their results, oldest to newest.](images/history.png)

The most recent event is at the top, so read the history from the bottom up to see each step in the process. Using this history, you can see exactly how your Workflow executed and pinpoint any places things might have gone wrong.

Temporal stores the results of each Activity in this history, as you can see in the image. If there was a system crash between the `getIP` and `getLocationInfo` Activity Executions, a new Worker would re-run the Workflow, but would use the previous Event History to reconstruct the Workflow's state. Instead of re-running the `getIP` function, it would use the previous run's value and continue on. This prevents duplicate executions. By relying on the stored Event History, Temporal ensures that the Workflow can recover seamlessly, maintaining reliability and consistency even after a crash.

In this application, this recovery isn't crucial. But imagine a situation where each Activity execution was a bank transaction. If a crash occurred between transactions, the Worker can pick up where the previous one failed. Nobody gets charged multiple times because something failed.

Next, you'll explore how Temporal handles failed Activities.

## Observe automatic retries

When you developed your Activities, you didn't include any error-handling code. So if there's a problem making the request, the Workflow will handle the error using the Retry Policy.

Test this out. Disconnect your local machine from the Internet by turning off your Wi-Fi connection or unplugging your network cable. In the playground, turn the **Network** switch below to **Off** instead.

<div class="network-switch"></div>

Then, with the local Temporal Service running and your Worker running, switch to the Terminal window where you ran your Workflow and run it again:

```command
mvn compile exec:java \
    -Dexec.mainClass="iplocate.IPLocateClient" \
    -Dexec.args="Ziggy" \
    -Dorg.slf4j.simpleLogger.defaultLogLevel=warn
```

In the playground, click **Run** below:

<div class="run-block" data-command="run" data-history data-name data-output-label="Client output"></div>

This time you don't get a response.

Visit `http://localhost:8233` to open the Temporal Web UI and locate the Workflow Execution that's currently running. In the playground, use the panel below.

<div class="webui"></div>

 When you select it, you'll see something like the following image, indicating that there's a problem:

![The timeline shows the Activity failure](images/timeline_failed.png)

As you can see, the `getIP` Activity has failed and Temporal is retrying it. Scroll down to the Event History and you'll see the failure represented there:

![The Event History shows the failed Activity](images/history_failed.png)
Select the **Pending Activity** item in the table to see why it failed and you'll see the stack trace:

![The Activity stack trace shows the error](images/activity_stack_trace.png)

In the **Last Failure** field, you can see that a DNS error occurred whilst resolving the domain name for the external service that the `getIP` activity fetches data from. In the playground, you'll see `Tunnel failed, got: 503` instead, because the playground blocks the request.

Connect to the internet again and wait. In the playground, turn the **Network** switch back to **On**. After a few moments, the Workflow recovers and completes:

![Once the network comes back the history updates and shows that things completed.](images/history_recovered.png)

If you return to your terminal where you launched the Workflow, you'll find your results there as well. In the playground, the result appears in the output below the **Run** button.

You can recover from failures by letting Temporal handle them for you instead of writing complex error-handling logic. You can also decide that you only want to retry a fixed number of times, or that you only want to recover on certain kinds of errors.

To try a fixed number of retries in the playground, turn the **Network** switch to **Off**, then click **Run with a five-attempt limit**. For this one run, the button removes the `//` in front of `.setMaximumAttempts(5)` in `GetAddressFromIPWorkflowImpl.java`, so the Workflow fails after five attempts. Turn the switch back to **On** when you're done.

<div class="network-switch"></div>

<div class="run-block" data-command="run" data-label="Run with a five-attempt limit" data-name data-output-label="Client output" data-variant-file="src/main/java/iplocate/GetAddressFromIPWorkflowImpl.java" data-variant-find="// .setMaximumAttempts(5)" data-variant-replace=".setMaximumAttempts(5)"></div>

## Conclusion

In this tutorial you built your first Temporal Application. You used the Temporal Java SDK to build a resilient application that recovered from failure. You wrote tests to verify that it works and reviewed the Event History for a working execution. You also tested your Workflow without an internet connection to understand how Temporal recovers from failures like network outages.

Take this application one step further and add a new Activity that gets the current weather for the location you found. The playground only allows calls to the two services this tutorial uses, so build that version on your own machine.

### Review

Answer the following questions to see whether you remember some important concepts from this tutorial:

<div class="quiz">
<h4>What are the four parts of a Temporal Workflow application?</h4>
<ol>
<li>A Workflow, a database, a message queue, and a client</li>
<li data-correct>A Workflow function, an Activity function, a Worker to host the Workflow and Activity code, and some way to start the Workflow</li>
<li>A Workflow, a Task Queue, the Web UI, and a Retry Policy</li>
<li>Activities, a scheduler, the Temporal Service, and a client</li>
</ol>
<p class="quiz-explanation">1. A Workflow function. 2. An Activity function. 3. A Worker to host the Workflow and Activity code. 4. Some way to start the Workflow.</p>
</div>

<div class="quiz">
<h4>How does the Temporal Server get information to the Worker?</h4>
<ol>
<li>It sends HTTP requests to an endpoint on the Worker</li>
<li>The client sends the work straight to the Worker</li>
<li data-correct>The Temporal Server adds Tasks to a Task Queue, and the Worker polls the Task Queue</li>
<li>It writes files that the Worker reads</li>
</ol>
<p class="quiz-explanation">The Temporal Server adds Tasks to a Task Queue, and the Worker polls the Task Queue.</p>
</div>
