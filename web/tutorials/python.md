---
id: hello-world-tutorial-python
title: Build a Temporal Application from scratch in Python
sidebar_position: 3
description: In this tutorial you will build a Temporal Application using the Python SDK. You'll write a Workflow, an Activity, tests, and define a Worker.
keywords: [python,temporal,sdk,tutorial, hello world, temporal application, workflow, activity]
last_update:
  author: Brian P. Hogan
  date: 2024-12-17
tags:
  - python
  - sdk
  - tutorial
image: /img/temporal-logo-twitter-card.png
---

<img className="banner" src="/img/sdk_banners/banner_python.png" alt="Temporal Python SDK" />

:::note Tutorial information

- **Level:** ⭐ Temporal beginner
- **Time:** ⏱️ ~15 minutes
- **Goals:** 🙌
  - Set up, build, and test a Temporal Application project from scratch using the [Python SDK](https://github.com/temporalio/sdk-python).
  - Identify the four parts of a Temporal Workflow application.
  - Describe how the Temporal Server gets information to the Worker.
  - Observe how Temporal recovers from failed Activities.

:::

### Introduction

Creating reliable applications is a difficult chore. [Temporal](https://temporal.io/) lets you create fault-tolerant, resilient applications using programming languages you already know, so you can build complex applications that execute reliably and recover from failures.

In this tutorial, you will build your first Temporal Application from scratch using the [Temporal Python SDK](https://github.com/temporalio/sdk-python). You'll develop a small application that asks for your name and then uses APIs to get the public IP address of the machine running the code and the location based on that address. External requests can fail due to rate limiting, network interruptions, or other errors. Using Temporal for this application will let you automatically recover from these and other kinds of failures without having to write explicit error-handling code.

The app will consist of the following pieces:

1. Two [Activities](https://docs.temporal.io/activities): Activities are functions called by Workflows, and they contain any logic that might fail or behave differently at different times. The first Activity will get your IP address, and the second Activity will use that IP address to get your location.

1. A [Workflow](https://docs.temporal.io/workflows): Workflows are functions that define the overall flow of the application.  Your Workflow will execute both Activities you define. It will call the first Activity to fetch the IP address, and then use the result of that Activity to call the next Activity to find your location.
2. A [Worker](https://docs.temporal.io/workers): Workers host the Activity and Workflow code and execute the code piece by piece.
3. A client program: You'll need to start your Workflow. In this tutorial you'll create a small program using the [Temporal Client](https://docs.temporal.io/encyclopedia/temporal-sdks#temporal-client) to start the Workflow.

You'll also write tests to ensure your Workflow executes successfully.

When you're done, you'll have a basic application and a clear understanding of how to build out the components you'll need in future Temporal applications in Python.

## Prerequisites

Before starting this tutorial:

- Follow the tutorial [Run your first Temporal application with the Python SDK](https://learn.temporal.io/getting_started/python/first_program_in_python/) to gain a better understanding of what Temporal is and how its components fit together.

:::note Playground
This playground runs a Temporal Service, the Temporal Web UI, and your Worker for you, so you don't need to install anything. When you're ready to build on your own machine, [set up a local development environment for developing Temporal applications using Python](https://learn.temporal.io/getting_started/python/dev_environment/).
:::

## Create a New Temporal Python Project

To get started with the Temporal Python SDK, you'll create a new Python project and a virtual environment, just like any other Python program you're creating. Then you'll add the Temporal SDK package to your project.

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

From the root of your new project directory, create a virtual environment and activate it:

```command
python -m venv env
source env/bin/activate
```

Then add the Temporal Python SDK as a project dependency, along with the `requests` package your Activities will use to make HTTP requests:

```command
python -m pip install temporalio requests
```
:::

With the project created, you'll create the application's core logic.

## Write functions to call external services

Your application will make two HTTP requests. The first request will return your current public IP, while the second request will use that IP to provide city, state, and country information.

You'll use Temporal Activities to make these requests. You use Activities in your Temporal Applications to execute [non-deterministic](https://docs.temporal.io/workflows#deterministic-constraints) code or perform operations that may fail, such as API requests or database calls.

If an Activity fails, Temporal can automatically retry it until it succeeds or reaches a specified retry limit. This ensures that transient issues, like network glitches or temporary service outages, don't result in data loss or incomplete processes.

Create the file `activities.py` in your project root:

```command
touch activities.py
```

Open the file `activities.py` in your editor and add the following code that imports the dependencies your Activities use:


<!--SNIPSTART python-ipgeo-activity-setup-->
[activities.py](activities.py)
<!--SNIPEND-->

With the Python SDK, you can define Activities as regular Python functions decorated with `@activity.defn`. You can also create them as methods on a class, which is necessary to pass shared objects like database connections. In this tutorial your Activities are functions that use the `requests` package, and you'll stub out the HTTP calls when you write tests for your Activities later.

Now add the following code to define a Temporal Activity that retrieves your IP address from `icanhazip.com`:

<!--SNIPSTART python-ipgeo-activity-ip-->
[activities.py](activities.py)
<!--SNIPEND-->

With the Temporal Python SDK, Activities can be regular functions or `async` functions. These Activities are regular functions because the `requests` package blocks while it waits for a response. You'll configure the Worker to run them on a thread pool later in this tutorial.

Like other Python functions, Activities report problems by raising exceptions. The call to `raise_for_status()` raises an exception if the service responds with an error status.

The response from `icanhazip.com` is plain-text, and it includes a newline, so you trim off the newline character before returning the result.

Notice that there's no error-handling code in this function. When you build your Workflow, you'll use Temporal's Activity Retry policies to retry this code automatically if there's an error.

Now add the second Activity that accepts an IP address and retrieves location data. In `activities.py`, add the following code to define it:

<!--SNIPSTART python-ipgeo-activity-location-->
[activities.py](activities.py)
<!--SNIPEND-->

This Activity follows the same pattern as the `get_ip` Activity.  This time, the service returns JSON data rather than text, so you use the response's `json()` method to parse the data.

While Activities can accept input arguments, it's a best practice to send a single argument rather than multiple arguments. In this case you only have a single String. If you have more than one argument, you should bundle them up in a serializable object. This is because later revisions to your Workflows and Activities that change the number of arguments sent to them can otherwise introduce the need for versioning. Review the [Activity parameters](https://docs.temporal.io/develop/python/activities/basics#activity-parameters) section of the Temporal documentation for more details.

You've created your two Activities. Now you'll coordinate them using a Temporal Workflow.

## Control application logic with a Workflow

Workflows are where you configure and organize the execution of Activities. You define a Workflow by writing a *Workflow Definition* using one of the Temporal SDKs.

In the Temporal Python SDK, a Workflow Definition is a class decorated with `@workflow.defn`. The class has one `async` method decorated with `@workflow.run`, which is the method Temporal calls to run the Workflow. That method can return a value, which you'll use to return the result of the Workflow Execution. Review the [Develop Workflows](https://docs.temporal.io/develop/python/workflows/basics#develop-workflows) section of the Temporal documentation for more about Workflows in Python.

Temporal Workflows [must be deterministic](https://docs.temporal.io/workflow-definition#deterministic-constraints) so that Temporal can replay your Workflow in the event of a crash. That's why you call Activities from your Workflow code. Activities don't have the same determinism constraints that Workflows have.

Create the file `workflows.py` in the root of your project:

```command
touch workflows.py
```

Then add the following code to import the Activities and configure how the Workflow should handle failures with a [Retry Policy](https://docs.temporal.io/encyclopedia/retry-policies).

<!--SNIPSTART python-ipgeo-workflow-imports-->
[workflows.py](workflows.py)
<!--SNIPEND-->


The Python SDK runs Workflow code in a sandbox that helps keep it deterministic. Importing your Activities inside `workflow.unsafe.imports_passed_through()` tells the sandbox to use the module that's already loaded instead of reloading it every time the Workflow runs.

With the imports and options in place, you can define the Workflow itself.

Add the following code to define the `GetAddressFromIP` Workflow, which will call both Activities, using the value of the first as the input to the second:

<!--SNIPSTART python-ipgeo-workflow-code-->
[workflows.py](workflows.py)
<!--SNIPEND-->

The `run` method accepts a string value that holds the name and returns a string value.  Like with Activities, you can send multiple inputs into a Workflow, but it's a good practice to combine those into a data class.

In this example, you have specified that the Start-to-Close Timeout for your Activities will be one minute, meaning that your Activity has one minute to complete before it times out. Of all the Temporal timeout options, `start_to_close_timeout` is the one you should always set.

You also set the Retry Policy for Activities. In this example, you're using the default Retry Policy values, so you don't need to specify the values, but by leaving them in you have a more clear picture of what happens. Note that the `maximum_attempts` is commented out, which means there's no limit to the number of times Temporal will retry your Activities if they fail.


This code does check for and handle errors, but because you've configured unlimited retries, there won't be any exceptions caught. However, if you change the Retry Policy's maximum retries, or you specify non-retryable exceptions, this code will be in place to handle those errors.

Next you'll create a Worker that executes the Workflow and Activities.

## Configure and run a Worker

When you start a Temporal Workflow, the Workflow and its Activities get scheduled on the Temporal Service's [Task Queue](https://docs.temporal.io/task-queue#task-queue). A [Worker](https://docs.temporal.io/workers) hosts Workflow and Activity functions and polls the Task Queue for tasks related to those Workflows and Activities. After the Worker runs the code, it communicates the results back to the Temporal Service where they're stored in the Event History. This records the Workflow's entire execution, enabling features like fault tolerance by allowing the Workflow to replay in case of Worker crashes or restarts.

You use the Temporal SDK to define a Worker Program.

In your Worker Program, you need to specify the name of the Task Queue, which must match the Task Queue name used whenever you interact with a Workflow from a client application. This ensures that the Worker processes tasks for the correct Workflow. The Task Queue name is a case-insensitive string. To ensure consistency and avoid errors, define the Task Queue name as a constant that can be reused throughout your code.

Create the file `shared.py`:

```command
touch shared.py
```

Open the file and add the following lines to the file to define the constant for the Task Queue:

<!--SNIPSTART python-ipgeo-shared-->
[shared.py](shared.py)
<!--SNIPEND-->

Now you can create the Worker program.

Create the file `worker.py` in your project root:

```command
touch worker.py
```

Then open `worker.py` in your editor and add the following code to define the Worker program:

<!--SNIPSTART python-ipgeo-worker-->
[worker.py](worker.py)
<!--SNIPEND-->

The code imports your Workflow and Activity Definitions and the Task Queue name. It then defines a `main` function that creates and runs a Worker.

You first create a client, and then you create a Worker that uses the client, along with the Task Queue it should listen on. The client connects to the Temporal Cluster running at `localhost` on port `7233`, and connects to the `default` namespace. You can change this by passing different values to `Client.connect`.

Then you register your Workflow and Activities with the Worker. Since your Activities are regular functions rather than `async` functions, the Worker needs an `activity_executor` to run them on, so you pass it a `ThreadPoolExecutor`.

In this case your Worker will run your Workflow and your two Activities, but there are cases where you could configure one Worker to run Activities, and another Worker to run the Workflows.

Now you'll start the Worker. Be sure you have started the local Temporal Service and execute the following command to start your Worker:

```command
python worker.py
```

The Worker runs without printing any output.

In the playground, you don't start the Worker yourself. It starts in the background every time you run the client, which works the same as leaving it running in its own terminal.

Your Worker is running and is polling the Temporal Service for Workflows to run, but before you start your Workflow, you'll write tests to prove it works as expected.

## Write tests to ensure things work

The Temporal Python SDK includes functions that help you test your Workflow executions. Let's add a basic unit test to the application to make sure the Workflow works as expected.

You'll use the [pytest](https://docs.pytest.org/) package to build your test cases and mock the Activity so you can test the Workflow in isolation.

Add the `pytest` and `pytest-asyncio` packages to your project by running the following command in your terminal:

```command
python -m pip install pytest pytest-asyncio
```

Now create a file to hold the test for your Workflow. Create a `tests` directory and the file `tests/test_workflows.py`:

```command
mkdir tests
touch tests/test_workflows.py
```

Add the following code to set up the testing environment:

<!--SNIPSTART python-ipgeo-workflow-test-setup-->
[tests/test_workflows.py](tests/test_workflows.py)
<!--SNIPEND-->

Add the following code to test the Workflow execution:

<!--SNIPSTART python-ipgeo-workflow-test-workflow-->
[tests/test_workflows.py](tests/test_workflows.py)
<!--SNIPEND-->

This test creates a test execution environment to run the Workflow.

Instead of using your actual Activities, you replace the Activities `get_ip` and `get_location_info` with mocks that return hard-coded values. This way you're testing the Workflow's logic independently of the Activities. If you wanted to test the Activities directly as part of an integration test, you would omit these mocks from the test.  Each mock uses `@activity.defn(name=...)` with the same name as the real Activity, so the Workflow calls the mock instead.

The test then executes the Workflow in the test environment and checks for a successful execution. Finally, the test ensures the Workflow's return value returns the expected value.

Ensure you've saved all your files and execute your tests with the following command:

```command
python -m pytest
```

The test environment starts, spins up a Worker, and executes the Workflow in the test environment. At the end, you'll see that your test passes:

```output
tests/test_workflows.py .                                                [100%]

============================== 1 passed in 0.79s ===============================
```

With a Workflow test in place, you can write unit tests for the Activities.

Both of your Activities make external calls to services that will change their results based on who runs them. It will be challenging to test these Activities reliably. For example, the IP address may vary based on your machine's location.

To ensure you can test the Activities in isolation, you’ll stub out the HTTP calls.

Create the file `tests/test_activities.py`

```command
touch tests/test_activities.py
```

Add the following code to import the testing libraries and Activities you'll use:

<!--SNIPSTART python-ipgeo-activity-test-setup-->
[tests/test_activities.py](tests/test_activities.py)
<!--SNIPEND-->


To ensure you don't make real HTTP requests, you use `unittest.mock.patch` to replace `requests.get` with a mock. When you write your tests you'll set what the mock returns so you can control the response.

Next, write the test for the `get_ip` Activity, using a mocked HTTP response to stub out actual HTTP calls so your tests are consistent. Notice that the mock response adds the newline character so it replicates the actual response:

<!--SNIPSTART python-ipgeo-activity-test-ip-->
[tests/test_activities.py](tests/test_activities.py)
<!--SNIPEND-->

The `patch` call replaces `requests.get` only inside the `with` block, so the change doesn't affect your other tests.

To test the Activity itself, you use an `ActivityEnvironment` to execute the Activity rather than directly calling the `get_ip` function. You get the result from the Activity Execution and then ensure it matches the value you expect.

To test the `get_location_info` Activity, you use a similar approach. Add the following code to the `tests/test_activities.py` file:

<!--SNIPSTART python-ipgeo-activity-test-location-->
[tests/test_activities.py](tests/test_activities.py)
<!--SNIPEND-->

This test looks similar to the previous test; you mock out the HTTP client and ensure it returns the expected data, and then you execute the Actiity in the test environment. Then you retrieve the value and ensure it's what you expect.

Run the tests again to see them pass. Use the `-v` option to see verbose output so you can see that all the tests ran:

```command
python -m pytest -v
```

```output
tests/test_activities.py::test_get_ip PASSED                             [ 33%]
tests/test_activities.py::test_get_location_info PASSED                  [ 66%]
tests/test_workflows.py::test_get_address_from_ip PASSED                 [100%]

============================== 3 passed in 0.77s ===============================
```

In the playground, click **Run tests** to run all three tests:

<div class="run-block" data-command="test" data-label="Run tests" data-output-label="pytest output"></div>

To see what a failing test looks like, this button changes the end of the expected greeting from `Planet Earth` to `Planet Mars` for one run:

<div class="run-block" data-command="test" data-label="Run tests with a wrong expected greeting" data-output-label="pytest output" data-variant-file="tests/test_workflows.py" data-variant-find="located in Planet Earth&quot;" data-variant-replace="located in Planet Mars&quot;"></div>

Now that you have your tests passing, it's time to start a Workflow Execution.

## Run the Workflow from a client

You can start a Workflow Execution by using the Temporal CLI or by writing code using the Temporal SDK.

Starting a Workflow Execution using the Temporal SDK involves connecting to the Temporal Server, configuring the Task Queue the Workflow should use, and starting the Workflow with the input parameters it expects. In a real application, you may invoke this code when someone submits a form, presses a button, or visits a certain URL. In this tutorial, you will create a small CLI program that runs your Temporal Workflow.

Create the file `client.py` in your project root:

```command
touch client.py
```

Open `client.py` in your editor and add the following code to the file to connect to the server and start the Workflow:

<!--SNIPSTART python-ipgeo-cli-client-->
[client.py](client.py)
<!--SNIPEND-->

In the `main` function you check to see if there is at least one argument passed and then capture the user's name from the arguments.

The `main` function then sets up a connection to your Temporal Server, invokes your Workflow, passes in an argument for the `name`, and assigns the Workflow a unique identifier using a UUID. The client dispatches the Workflow on the same Task Queue that the Worker is polling on. That's why you used a constant to ensure the Task Queue name is consistent. If there's a mismatch, your Workflow will execute on a different Task Queue and there won't be any Workers polling for tasks.

:::note Specify a Workflow Id

A Workflow ID is unique in a Namespace and identifies a Workflow Execution. Using an identifier that reflects some business process or entity is a good practice. For example, you might use a customer identifier as part of the Workflow Id if you run one Workflow per customer. This would make it easier to find all Workflow Executions related to that customer later.

In this tutorial you're generating a UUID and appending it to a string that identifies the Workflow.
:::

You can [get the results](https://docs.temporal.io/develop/python/client/temporal-client#get-workflow-results) from your Workflow right away, or you can get the results at a later time. This implementation attempts to get the results immediately by calling `execute_workflow`, which waits until the Workflow Execution completes and returns its result.

Now you can run your Workflow. First, ensure that your local Temporal Service is running, and that your Worker program is running also.

Then open a new terminal and switch to the project directory:

```command
cd temporal-ip-geolocation
```

Now run the following command to run the Workflow using the client program you wrote:

```command
python client.py Ziggy
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

Temporal stores the results of each Activity in this history, as you can see in the image. If there was a system crash between the `get_ip` and `get_location_info` Activity Executions, a new Worker would re-run the Workflow, but would use the previous Event History to reconstruct the Workflow's state. Instead of re-running the `get_ip` function, it would use the previous run's value and continue on. This prevents duplicate executions. By relying on the stored Event History, Temporal ensures that the Workflow can recover seamlessly, maintaining reliability and consistency even after a crash.

In this application, this recovery isn't crucial. But imagine a situation where each Activity execution was a bank transaction. If a crash occurred between transactions, the Worker can pick up where the previous one failed. Nobody gets charged multiple times because something failed.

Next, you'll explore how Temporal handles failed Activities.

## Observe automatic retries

When you developed your Activities, you didn't include any error-handling code. So if there's a problem making the request, the Workflow will handle the error using the Retry Policy.

Test this out. Disconnect your local machine from the Internet by turning off your Wi-Fi connection or unplugging your network cable. In the playground, turn the **Network** switch below to **Off** instead.

<div class="network-switch"></div>

Then, with the local Temporal Service running and your Worker running, switch to the Terminal window where you ran your Workflow and run it again:

```command
python client.py Ziggy
```

In the playground, click **Run** below:

<div class="run-block" data-command="run" data-history data-name data-output-label="Client output"></div>

This time you don't get a response.

Visit `http://localhost:8233` to open the Temporal Web UI and locate the Workflow Execution that's currently running. In the playground, use the panel below.

<div class="webui"></div>

 When you select it, you'll see something like the following image, indicating that there's a problem:

![The timeline shows the Activity failure](images/timeline_failed.png)

As you can see, the `get_ip` Activity has failed and Temporal is retrying it. Scroll down to the Event History and you'll see the failure represented there:

![The Event History shows the failed Activity](images/history_failed.png)
Select the **Pending Activity** item in the table to see why it failed and you'll see the stack trace:

![The Activity stack trace shows the error](images/activity_stack_trace.png)

In the **Last Failure** field, you can see that a DNS error occurred whilst resolving the domain name for the external service that the `get_ip` activity fetches data from. In the playground, you'll see a `ProxyError` with `503 Service Unavailable` instead, because the playground blocks the request.

Connect to the internet again and wait. In the playground, turn the **Network** switch back to **On**. After a few moments, the Workflow recovers and completes:

![Once the network comes back the history updates and shows that things completed.](images/history_recovered.png)

If you return to your terminal where you launched the Workflow, you'll find your results there as well. In the playground, the result appears in the output below the **Run** button.

You can recover from failures by letting Temporal handle them for you instead of writing complex error-handling logic. You can also decide that you only want to retry a fixed number of times, or that you only want to recover on certain kinds of errors.

To try a fixed number of retries in the playground, turn the **Network** switch to **Off**, then click **Run with a five-attempt limit**. For this one run, the button removes the `#` in front of `maximum_attempts=5` in `workflows.py`, so the Workflow fails after five attempts. Turn the switch back to **On** when you're done.

<div class="network-switch"></div>

<div class="run-block" data-command="run" data-label="Run with a five-attempt limit" data-name data-output-label="Client output" data-variant-file="workflows.py" data-variant-find="# maximum_attempts=5," data-variant-replace="maximum_attempts=5,"></div>

## Conclusion

In this tutorial you built your first Temporal Application. You used the Temporal Python SDK to build a resilient application that recovered from failure. You wrote tests to verify that it works and reviewed the Event History for a working execution. You also tested your Workflow without an internet connection to understand how Temporal recovers from failures like network outages.

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
