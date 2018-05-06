/*
  ==============================================================================

    This file was auto-generated!

    It contains the basic framework code for a JUCE plugin editor.

  ==============================================================================
*/

#include "PluginProcessor.h"
#include "PluginEditor.h"

//==============================================================================
FmfmAudioProcessorEditor::FmfmAudioProcessorEditor (FmfmAudioProcessor& p)
    : AudioProcessorEditor (&p), processor (p), rootComponent (p)
{
    // Make sure that before the constructor has finished, you've set the
    // editor's size to whatever you need it to be.
    setSize (400, 300);

    addAndMakeVisible (rootComponent);
    rootComponent.setSize (this->getWidth(), this->getHeight());
}

FmfmAudioProcessorEditor::~FmfmAudioProcessorEditor()
{
}

//==============================================================================
void FmfmAudioProcessorEditor::paint (Graphics& g)
{
    g.fillAll (getLookAndFeel().findColour (ResizableWindow::backgroundColourId));
    g.setColour (Colours::white);
    g.setFont (15.0f);
    g.drawFittedText ("test", getLocalBounds(), Justification::centred, 1);
}

void FmfmAudioProcessorEditor::resized()
{
    // This is generally where you'll want to lay out the positions of any
    // subcomponents in your editor..

    rootComponent.setSize (this->getWidth(), this->getHeight());
}
