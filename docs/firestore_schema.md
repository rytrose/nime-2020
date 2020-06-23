# Firestore Schema

This document defines the fields and expected values in the Firestore collections/documents for the installation.

## `users` collection

The `users` collection maintains a document for each Firebase user, by their user ID. These users are anonymous, and are created when a user firsts visits or clears their browser history or cookies.

Each document in `users` contains a `lastOperations` collection, which contains documents with and ID of a room name. Each document in the `lastOperations` collection for a user contains one field `lastOperation` which stores the timestamp of the last operation submitted by that user in that room. Here is an example JSON representation the `users` collection:

```javascript
{
    "users": {
        "<userID>": {
            "lastOperations":  {
                "<roomName>": {
                    "lastOperation": 1592172062716
                },
                "<roomName2>": {
                    "lastOperation": 1592172071283
                },
            }
        }
    }
}
```

## `rooms` collection

The `rooms` collection contains data for each "wall" we're presenting to users, including general metadata and behaviors for that room.

Each document in `rooms` uses the room name as the ID, and contains the following fields:

| Property         | Type    | Description                                                                                        | Enum        |
|------------------|---------|----------------------------------------------------------------------------------------------------|-------------|
| `active`         | boolean | Flag for whether or not this room should be shown to users.                                        |             |
| `type`           | string  | An enumeration for the type of interface the room presents to the user.                            | `PIANO_ROLL_SEQUENCER` |
| `hash`           | string  | The hash used to load the necessary assets for the room.                                           |             |
| `thumbnail`      | string  | A URL to an image to use on the landing page representing the room.                                |             |
| `description`    | string  | User facing description of the room.                                                               |             |
| `actionsAllowed` | number  | The number of actions a user can take in this room before submitting and waiting `actionWaitTime`. |             |
| `actionWaitTime` | number  | The duration in milliseconds between action submissions a user must wait.                          |             |
| `rules`          | string  | A JSON string that can be interpreted by the client to enforce rules in the room.                  |             |

The value of `rules` is a JSON array that contains objects that follow this schema:

| Property     | Type   | Description                                                                       | Enum                                                                  |
|--------------|--------|-----------------------------------------------------------------------------------|-----------------------------------------------------------------------|
| `ruleType`   | string | An enumeration for the type of rule, used to determine how to parse `ruleParams`. | `STEPS_DISABLED`, `NOTES_DISABLED`, `KNOBS_DISABLED`, `KNOBS_MIN_MAX` |
| `ruleParams` | string | A JSON string that defines how the rule should be applied.                        |                                                                       |

The following describes the `ruleParams` for each `ruleType`:

| `ruleType`       | Property       | Type          | Description                                                    |
|------------------|----------------|---------------|----------------------------------------------------------------|
| `STEPS_DISABLED` | `steps`        | array[number] | An array of step indices to disable.                           |
| `NOTES_DISABLED` | `notes`        | array[string] | An array of note names to disable (e.g. "C4", "A#2").          |
| `KNOBS_DISABLED` | `knobs`        | array[string] | An array of knob IDs to disable.                               |
| `KNOBS_MIN_MAX`  | `knobs`        | array[object] | An array of objects describing knob labels and min/max values. |
|                  | `knobs[i].id`  | string        | The knob ID this refers to.                                    |
|                  | `knobs[i].min` | number        | The minimum value this knob will allow.                        |
|                  | `knobs[i].max` | number        | The maximum value this knob will allow.                        |

### The `ruleFunction` property

An additional, optional property of a `rules` object is `ruleFunction`. This property contains a serialized function that expects a first argument of `t`, the current `Date.now()` timestamp, and a second argument `ruleParams`, the current value of `ruleParams` for that rule. The function should return a full, new `ruleParams` object for that rule. The `ruleParams` must also contain a `ruleFunctionInterval` property that determines how frequently the `ruleFunction` should be called.
