# Reaper OSC Action

This is a Elgato StreamDeck Plugin to send Command IDs to the DAW Reaper via OSC. It is written in Go.

Demo Video can be found here: https://www.youtube.com/watch?v=DTwFpP6xsbU

## Building

Building the Plugin is quite easy:

`make`

## Installation

Symlink the plugin folder to the StreamDeck plugins folder

`ln -s ./org.smyck.reaper_osc_action.sdPlugin ~/Library/Application\ Support/com.elgato.StreamDeck/Plugins/`

Restart the StreamDeck app and verify in the Preferences / Plugins tab that
the "Reaper OSC Action" plugin appears.

Then place it on one of the StreamDeck Buttons and add "127.0.0.1" as IO, the
port that was configured in Reaper and a command id of your choice.
