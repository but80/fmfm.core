/*
  ==============================================================================

    This file was auto-generated!

    It contains the basic framework code for a JUCE plugin editor.

  ==============================================================================
*/

#pragma once

#include "../JuceLibraryCode/JuceHeader.h"
#include "PluginProcessor.h"
#include "RootComponent.h"

//==============================================================================
/**
*/
class FmfmAudioProcessorEditor  : public AudioProcessorEditor
{
public:
    FmfmAudioProcessorEditor (FmfmAudioProcessor&);
    ~FmfmAudioProcessorEditor();

    //==============================================================================
    void paint (Graphics&) override;
    void resized() override;

private:
    // This reference is provided as a quick way for your editor to
    // access the processor object that created it.
    FmfmAudioProcessor& processor;

    RootComponent rootComponent;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (FmfmAudioProcessorEditor)
};
