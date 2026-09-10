# Device acceptance trial

The software POC is tested at 800 × 480 and 1280 × 800 with Fyne's software test renderer. This verifies layout and interaction without proving physical touch or GPU behavior. Complete these steps on the chosen Linux device and waterproof enclosure before a lesson trial.

1. Build and install the Go Linux archive for the device's architecture, with the required OpenGL/X11 libraries. Verify normal and fullscreen launch, readable font rendering, and Fyne import/export dialogs. Confirm physical control sizes at the display's actual DPI.
2. Import synthetic class data, then disable networking. Open each class, change attendance, choose a skill in Skills, then choose a swimmer and tap the completed subskills. Confirm the checklist stays open after the last subskill; save with the green checkmark rail button. Repeat until three saved successful attempts complete a skill; verify feedback at 0/3, 1/3, 2/3, and completion.
3. Leave a subskill checklist unfinished, save it with the green checkmark, and verify the attempt appears in Feedback without increasing successful progress. Start another attempt and discard it with the red X. Hold a checked skill for two seconds and confirm its newest successful check and observation are removed. Review an older attempt through Feedback and confirm corrections still change earned progress only after submission.
4. With synthetic data, interrupt device power after a completed save and verify database integrity and saved work after boot. The automated suite tests abrupt process exit; real storage and power-loss behavior needs the hardware check.
5. Re-import a reordered roster with stable IDs and confirm that all results still belong to the correct swimmer. Export and inspect the snapshot on a separate computer.
6. Try the actual enclosure and input device with wet hands and representative water exposure. Check accidental touches, button activation, paging, visibility, and recovery from a mistaken mark. Software cannot establish waterproofing or reliable submerged touch response.
7. Run for a full representative teaching shift on battery. Check startup time, heat, readability in pool lighting, touch latency, and endurance. Record the exact hardware, OS, screen resolution/DPI, Fyne version, enclosure, and results before considering the wearable trial passed.

Actual hardware selection, waterproofing, wet-touch reliability, battery endurance, and automatic data transfer remain outside the completed software checks.
