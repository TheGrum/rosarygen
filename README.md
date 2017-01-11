#  RosaryGen  #
===============

[![GoDoc](https://godoc.org/github.com/TheGrum/rosarygen?status.svg)](https://godoc.org/github.com/TheGrum/rosarygen)

Install:
```
go get github.com/TheGrum/rosarygen
```

=====

A specialized tool to produce Rosary audio files from a collection of recorded prayers.

This is meant to solve a particular problem. If you just listen to the rosary on your drive home, buying one that you enjoy is usually manageable. 

On the other hand, if you are *using* a recorded rosary as an aid to praying the Rosary yourself, then you may have found that commercially available and public-domain recordings share a set of faults.

They may be:

  * Too fast
  * Too slow
  * Missing preferred additions (Fatima prayers, Flame of Love prayers, Papal additions, etc).
  * Wrong order of additions
  * Pronunciations from other dialects (Hal-lowed vs Hal-low-ed, etc)
  * No time given for meditating on mystery
  * Intro/Outro/Background music
  * Too little breathing space between prayers
  * Too much breathing space between prayers
  * Breathing pattern different from desired
  * Missing the Luminous mysteries
  * Inconsistent (some decades end with Oh My Jesus, others don't, etc.)

RosaryGen supports describing a desired structure of prayers (including adding prayers RosaryGen does not know about out of the box), a desired set of mysteries, specifying several options for selecting how a prayer may be divided into sections (for call/response, etc), and a template for the output filenames.

In turn, it will take this and provide a list of needed audio files (searching a list of directories in the provided order, making it simple to layer customized audio files over a standard set), and when all files are available, will stream the input files into the appropriate set of output files.

### Usage

`rosarygen Render`

```
Usage of rosarygen:

rosarygen [parameters] command

  -config string
    	Path to ini config for using in go flags. May be relative to the current executable path.
  -configUpdateInterval duration
    	Update interval for re-reading config file set via -config flag. Zero disables config file re-reading.
  -dumpflags
    	Dumps values for all flags defined in the app into stdout in ini-compatible syntax and terminates the app.
  -format string
    	wav or flac (default "wav")
  -gapLength int
    	tenths of seconds of silence to add between prayers (default 5)
  -groups string
    	Mystery decade groupings to generate. Possible values: All, Old (All excluding Luminous), Joyful, Luminous, Sorrowful, Glorious, and Custom (specify list of mysteries with mysteries) (default "All")
  -idirs string
    	Comma separated list of audio data folders, searched in order given (default "data")
  -mysteries string
    	List of mysteries to use in place of group. Use ListMysteries to see options.
  -odir string
    	output folder (default "output")
  -ofilename string
    	Output filename template. Available fields: Group, GroupNum, Mystery, MysteryNum, Prayer, PrayerNum, OutputFileNum, XthGroupMystery (default "{{.GroupNum}} {{.Group}} Mysteries")
  -structure string
    	Rosary structure to use. Use ListStructures to see options. (default "basic")
```

### Commands

 * PrintOptions - This prints the selected prayer options from the `options.toml` file.

 * ListPrayers - lists all available prayers from the prayers.toml and options.toml file (use to verify that added prayers are being picked up)

 * ListGroups - lists the groups of mysteries that can be specified in the relevant parameters

 * ListMysteries - lists the mysteries that can be specified in the relevant parameters 

 * ListStructures - lists the rosary structures the program knows about (use to verify that added structures are being picked up)

Commands below this point actually apply the selected rosary structure, and operate on the result.

 * Prayers - lists the actual prayers specified by the combination of -structure, -group, and -mysteries, in the order they will be output.

 * MissingFiles - does a dry run of the rendering and reports any files that have no matching file in any of the specified -idirs

 * ActualFiles - does a dry run and reports all matched files - use to verify that options are being chosen correctly and that idirs are having the desired effect

 * Render - this is the real deal - opens each actual file in turn, streaming them into the final output file[s]

 * RenderList - even 'realer' deal - if you are composing an MP3 CD as opposed to an Audio one, a rosary will fill only a tiny portion of it. RenderList lets you prepare a file to feed RosaryGen to produce multiple rosaries, chaplets, and prayers to fill such a CD.

### Filename Template

The output filename parameter `-ofilename` uses Go Templates, which need a little explanation. Basically, it uses embedded values of the form `{{.FieldName}}` that get replaced with values that reflect the state of the output.

## Examples:

Default, suitable for one file per group of mysteries (fine for mp3 players/phones):
-ofilename "{{.GroupNum}} {{.Group}} Mysteries"

 * 1 Preamble Mysteries.wav
 * 2 Joyful Mysteries.wav
 * . . .
 * 6 Postamble Mysteries.wav

Suitable for CD Track titles:
-ofilename "{{.OutputFileNum | printf \"%02d\"}} {{.XthGroupMystery}}" or -ofilename "{{.ZeroNum .OutputFileNum 2}} {{.XthGroupMystery}}" or -ofilename "{{.CDTrack}}

 * 01 Preamble.wav
 * 02 First Joyful Mystery.wav
 * 03 Second Joyful Mystery.wav
 * etc...

## Available Fields:

 * Group - "Preamble", "Postamble", "Joyful", "Luminous", "Sorrowful", "Glorious"
 * DecadeNumWord - "First", "Second", "Third", etc - which mystery in this group are we on
 * Mystery - "Transfiguration", "Scourging", etc...
 * PrayerName - prayer name including spaces, commas, etc.
 * OutputFileNum - this cannot itself trigger a change in output file, but when a change occurs, it is incremented
 * GroupNum - counts up
 * MysteryNum - this is the mystery number from the configuration file, NOT a counting number. It can be out of order.
 * PrayerNum - counts up
 * HailMaryNum - counts Hail Mary's within a Mystery
 * XthGroupMystery - function that returns the commonly used 'First/Second/Third/etc Joyful/Sorrowful/etc Mystery' form of name.
 * XofGroup - "Preamble/First Of Five/Postamble" - useful when using groups one/two/three/four/five,etc
 * CDTrack - provides an easier CD track title using a file number formatted with 2 digits including a leading zero so that alphabetical sorting gives the proper order. If you need 3 or more digits, see the CD Track example above for the in-template way of doing this, where you can change the 02 to the desired number of digits.
 

## Effect

When the filename resulting from applying the current state to the template changes, the OutputFileNum is incremented, the name is recalculated again, and the previous file is closed and the new one opened.

So the filename template *directly* controls how many files are produced, and how fine grained they are. There is no other option to say you want one file per group of mysteries, or one per mystery. If the filename would differ between mysteries you will get one file per mystery. A simple filename with no template entries will produce a single monolithic output.

### Options

RosaryGen expects an options.toml file containing an [options] section with one entry per prayer that has options, and a numeric value selecting which option to use.

```
[options]
 ourfather = 1
 hailmary = 4
 glorybe = 3
 meditation = 1
 hailholyqueen = 1
```

A missing option will be interpreted as a 1.

The options.toml file may *also* contain [prayer] and [structure] entries - see prayers.toml and structures.toml for examples. An entry in the options.toml will replace an identically keyed entry in the prayers.toml or structures.toml file.

Filenames defined on prayers are *also* templated on the same running status used for the output filename. This is particularly relevant for defining audio files that need to differ for each mystery (such as announcing the mystery, or a meditation for a mystery, etc), and examples of this may also be found in the prayers.toml file.

### RenderList

Call with rosarygen RenderList filename, or pipe into rosarygen RenderList. OutputFileNums will increment continually across all rendered files.

The file format understands four line types:

 * Parameter setting - has '=' in it somewhere, parameters are same as on command line, with one addition - filenum will override the current filenum

```idirs=rosary/basic,rosary/extended,rosary/extra,rosary/chaplets gap=3 odir=test ofilename={{.CDTrack}} structure=extended
 ```

If structure= is present, this will Render that structure to the disk. Spaces in parameters must be escaped with '\\'.

```ofilename={{.ZeroNum\ .OutputFileNum\\ 2}}\ {{.PrayerName}}```

 * Rendering - has neither '=' nor '\[', may have a comma - always in the form 'structure' or 'structure,groups'. As a special bit of magic for RenderList, if structure is the name of a prayer instead of a structure, it will dynamically create a structure with that single prayer in the preamble and Render it.

```basic,all```
```extended,old```

 * Literal structure - has '[', no '=' - dynamically creates a structure and then renders it

```[signofthecross,ourfather,hailmary,glorybe,signofthecross]```

 * Comments - start with '#', are ignored

## Example RenderList

```
idirs=example,rosary/basic,rosary/extended,rosary/extra,rosary/chaplets gap=3 odir=test ofilename={{.CDTrack}}
structure=example
ofilename={{.ZeroNum\ .OutputFileNum\ 2}}\ Chaplet\ of\ Divine\ Mercy\ {{.XofGroup}}
chapletofdivinemercy,five
ofilename={{.ZeroNum\ .OutputFileNum\ 2}}\ Chaplet\ of\ St.\ Michael
chapletofstmichael,none
ofilename={{.ZeroNum\ .OutputFileNum\ 2}}\ {{.PrayerName}}
memorare
magnificat
divinepraises
ohmary
queenofheavenrejoice
queenoftheholyrosary
augustqueenofheaven
comeohcreatorspirit
hailbrightstarofocean
ofilename={{.ZeroNum\ .OutputFileNum\ 2}}\ Common\ Setting
[signofthecross,ourfather,hailmary,glorybe,signofthecross]
```

## Produces:

```
File test/01 Preamble.wav written.
File test/02 First Joyful Mystery.wav written.
File test/03 Second Joyful Mystery.wav written.
File test/04 Third Joyful Mystery.wav written.
File test/05 Fourth Joyful Mystery.wav written.
File test/06 Fifth Joyful Mystery.wav written.
File test/07 First Luminous Mystery.wav written.
File test/08 Second Luminous Mystery.wav written.
File test/09 Third Luminous Mystery.wav written.
File test/10 Fourth Luminous Mystery.wav written.
File test/11 Fifth Luminous Mystery.wav written.
File test/12 First Sorrowful Mystery.wav written.
File test/13 Second Sorrowful Mystery.wav written.
File test/14 Third Sorrowful Mystery.wav written.
File test/15 Fourth Sorrowful Mystery.wav written.
File test/16 Fifth Sorrowful Mystery.wav written.
File test/17 First Glorious Mystery.wav written.
File test/18 Second Glorious Mystery.wav written.
File test/19 Third Glorious Mystery.wav written.
File test/20 Fourth Glorious Mystery.wav written.
File test/21 Fifth Glorious Mystery.wav written.
File test/22 Postamble.wav written.
File test/23 Chaplet of Divine Mercy Postamble.wav written.
File test/24 Chaplet of Divine Mercy First Of Five.wav written.
File test/25 Chaplet of Divine Mercy Second Of Five.wav written.
File test/26 Chaplet of Divine Mercy Third Of Five.wav written.
File test/27 Chaplet of Divine Mercy Fourth Of Five.wav written.
File test/28 Chaplet of Divine Mercy Fifth Of Five.wav written.
File test/29 Chaplet of Divine Mercy Postamble.wav written.
File test/30 Chaplet of St. Michael.wav written.
File test/31 Memorare.wav written.
File test/32 Magnificat.wav written.
File test/33 The Divine Praises.wav written.
File test/34 Oh, Mary.wav written.
File test/35 Queen of Heaven, Rejoice.wav written.
File test/36 Queen of the Holy Rosary.wav written.
File test/37 August Queen of Heaven.wav written.
File test/38 Come Oh Creator Spirit Blest.wav written.
File test/39 Hail, Bright Star of Ocean.wav written.
File test/40 Common Setting.wav written.
```

### Caveats/Known Problems

The library being used only supports FLAC decoding, not encoding, so currently FLAC is disabled. 

WAV files are assumed to be 48000khz - if this is not the case, they should be resampled to that rate prior to use. If output contains chipmunk noises where you expect a prayer, this is the probable culprit. 

WAV files are assumed to be stereo. If you hear the prayer expected, but it sounds sped up, this is a possible culprit.

### Structure

Depending on your desires, you may not need to touch structures. Three have been provided; one for the traditional simple rosary ('basic'), a slightly extended form, and a Fatima form that includes the prayers added at Fatima.

If you do need to implement your own structure, copying one of the existing ones into options.toml and editing it is the easiest way to proceed. Each structure is simply a collection of lists of prayers:

 * the preamble, which precedes the first decade
 * the group, which precedes/announces a set of mysteries (i.e. "The Joyful Mysteries") and is repeated for each group of mysteries
 * the mystery, which is repeated once for each mystery
 * the postamble, which is the set of prayers that ends the rosary

Any of these groups may be empty; for example, if you were recording your own mystery meditations, you could produce a test file of all of them to listen to with a structure like:

```
[structure.mysteries]
 name = "Mysteries"
 preamble = []
 group = [ "announcegroup" ]
 mystery = [ "announcemystery", "meditation" ]
 postamble = []
```

### Where do I get the audio files to make a rosary?

 There are several worthwhile options:

 * Record your own

Apply a desired structure and use MissingFiles to report what files need to be recorded for your chosen result. A bathroom or a car can provide a quiet environment for recording, while most modern cell-phones have voice recording functions that are suitable. You may need to run the results through a conversion program to get the 48000khz stereo WAV file needed. Recording multiple prayers and takes into a single file and cutting them in an audio editor such as Audacity may be easier than attempting to cleanly start and stop a phone recording.

 * Fix your favorite

If you have a favorite recording of the Rosary that is merely lacking in one of the respects listed above, you can obtain the audio file (either by downloading a file, or ripping a cd/dvd), then use an audio editor to cut out individual prayers and save them. Scan through the file, find the cleanest recording of each prayer to save, and RosaryGen will stitch them all back together, improving consistency.

 * Both at once

Use the -idirs function. Save the audio files from your favorite recording in one folder (or maybe make a pastiche from multiple favorite recordings!), and in a different folder, record the files that MissingFiles still reports, filling in the holes in your favorite recording. 

 * From me

I have recorded a fairly complete set of prayers that might be said with or in a rosary that I will eventually get uploaded. They can be used to render a complete rosary (assuming you can stand my accent!), or used to backstop your own recordings or favorites.

Uploaded - Hopefully this link will work: [RosaryGen Audio Files](https://drive.google.com/drive/folders/0Bx9pm0YfmfufTEZSVHh1QmFfQ1k?usp=sharing)

I divided them into three sets - basic is the prayers and elements needed for the basic structure. basic + extended has everything needed for the extended structure. Extra has additional recordings. 
