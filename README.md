# demoverse

Demoverse records demonstrations for [µniverse](https://github.com/unixpickle/muniverse). In other words, it records you playing games so that AI can learn to copy you.

# Requirements

A current version of `go` installed on your system

# Installation

Clone this repository and run `$ go install` inside the cloned folder. Afterwards you should be able to run `$ demoverse`. 

# Usage
Run `$ demoverse` in a directory you want to create the recordings in. A folder called *recordings* will automatically be created and contains recordings of you playing the games which can then be used for training your agent.

Open `localhost:8080` in your browser and you will see a list of available games. Just click on any game to start playing. The recording will automatically be started once your playing the game.

For the full list of available parameters for **demoverse** run:
```
$ demoverse -help

Usage of demoverse:
  -addr string
    	address to listen on (default ":8080")
  -assets string
    	asset directory path (default "assets")
  -cursor
    	render cursor
  -filter value
    	event filter (NoFilter or DeltaFilter)
  -frametime duration
    	time per frame (default 100ms)
  -gamesdir string
    	custom games directory
  -image string
    	custom docker image
  -outdir string
    	recordings directory (default "recordings")
  -templates string
    	template directory path (default "templates")
```

# Clone a policy

The **muniverse-agent** tool can be used to learn a policy from the recordings created with **demoverse**.

Therefore split the recordings created with *demoverse* in two sets. One subset is used for training, the second one will be used for validation.

Below you can see an example for the game *BirdyRush-v0*:
```
$ muniverse-agent clone -dir ./recordings/BirdyRush-v0/ -env BirdyRush-v0 -out birdyrush_policy -validation ./validation/BirdyRush-v0/
```
The `clone` flag is used to clone a policy from demonstrations.

`-dir`: the directory containing the recordings created by **demoverse**

`-env`: the environment for which a policy should be cloned 

`-out`: the name of the policy 

`-validation`: the directory containing the validation subset of the recordings

Once the policy is cloned from demonstrations, the policy can be used by the **muniverse-agent**.





