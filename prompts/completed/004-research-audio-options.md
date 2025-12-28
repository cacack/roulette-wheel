<research_objective>
Research and document the best options for implementing realistic roulette wheel audio in this Go/Ebitengine game. The goal is to determine the most practical approach for:
1. A continuous rolling/rumbling sound while the ball orbits the outer wooden track
2. Distinct bounce/ping sounds when the ball drops and bounces on the numbered metal slots
3. Perfect synchronization between audio and ball physics

Thoroughly explore multiple approaches and provide actionable recommendations.
</research_objective>

<context>
This is a Go game using Ebitengine for graphics and audio. The current audio implementation is in `./audio/audio.go` and uses procedurally generated sounds.

Examine the current implementation:
- `./audio/audio.go` - Current audio system
- `./ball/ball.go` - Ball physics (phases, callbacks for tick/bounce/settle)
</context>

<scope>
Research these approaches:

1. **Procedural Audio Generation**
   - Can we synthesize realistic rolling/rumbling sounds mathematically?
   - What techniques exist (noise + filtering, granular synthesis)?
   - Pros/cons for this use case

2. **Embedded Audio Samples**
   - What free/CC0 sound libraries have roulette/casino sounds?
   - How to embed WAV/MP3 files in Go binaries?
   - File size implications
   - How to loop rolling sounds seamlessly?

3. **Hybrid Approach**
   - Procedural for some sounds, samples for others
   - Which sounds benefit most from each approach?

4. **Ebitengine Audio Capabilities**
   - What audio formats does Ebitengine support natively?
   - How to play looping background sounds?
   - How to trigger one-shot sounds with precise timing?
   - Can we vary pitch/volume dynamically based on ball speed?

5. **Synchronization Strategy**
   - How to sync rolling sound intensity with ball angular velocity?
   - How to ensure bounce sounds trigger at exact moment of collision?
</scope>

<deliverables>
Save findings to: `./research/audio-options.md`

Structure the document as:
1. **Executive Summary** - Recommended approach in 2-3 sentences
2. **Option Analysis** - Each approach with pros/cons/feasibility
3. **Sound Sources** - Specific libraries, URLs, license info if using samples
4. **Technical Implementation** - How to integrate with current audio.go
5. **Recommendation** - Detailed implementation plan for chosen approach
</deliverables>

<evaluation_criteria>
- Realism: Which approach sounds most like a real roulette wheel?
- Simplicity: Can we implement this without major refactoring?
- Binary size: Embedded samples shouldn't bloat the executable excessively
- Licensing: Any samples must be free for use (CC0, public domain, or permissive license)
</evaluation_criteria>

<verification>
Before completing, verify:
- All five research areas are addressed
- Specific, actionable recommendations are provided
- Any referenced sound libraries include URLs and license info
- Implementation plan is concrete enough to execute as a follow-up task
</verification>
