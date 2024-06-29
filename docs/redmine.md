# Redmine integration

Honoroit can optionally integrate with Redmine to provide 2-way communication between Redmine and Matrix,
allowing you to manage requests from Redmine directly in Matrix.

## Configuration

On Honoroit side: follow the redmine's env vars in the [main README](../README.md#redmine).

On Redmine side: no additional configuration is required, just create a project you want to use with Honoroit.


## Usage

Once a new request is created in Honoroit, a new issue will be created in Redmine.
Any message sent in the operators thread and users room will be added as a comment to the Redmine issue.

Any note added to the Redmine issue will be sent to the operators thread and users room.
Private notes will be sent only to the operators thread.

When the request is closed in Honoroit, the Redmine issue will be closed as well.
When the request is closed in Redmine, the thread will be closed in Honoroit.
