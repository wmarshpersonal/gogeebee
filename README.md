
# gogeebee

This is an implementation of a cycle-accurate Game Boy emulator in Go.

## CPU
The CPU emulation is designed in terms of the individual units of the real CPU. Instead of writing code to emulate opcodes as a whole, individual equivalents of the `ALU`, `IDU`, data bus etc. operations are implemented, and opcodes defined in a data-driven way (via YAML files). For example, the `ADD N` opcode:

```yaml
add n:
  code: 0xC6
  cycles:
    - addr: PC
      data: Z ←
      idu:  ++
    - addr: PC
      alu:  A ← A + Z
      ftch: YES
```
1) The first cycle sets the address bus to the value of the `PC` register, and the data bus is instructed to read from that address to the internal `Z` register. The `IDU` increment operation works on the register selected into address bus.
2) The second cycle has the `ALU` perform addition between `A` & `Z`, and instructs the CPU core to perform a fetch from `PC`.

This matches quite accurately how the real CPU works, and as long as every operation is implemented the opcodes can be defined in data, rather than code. Cycle accuracy is easier to achieve since each cycle matches what the CPU is actually doing during that cycle. There's also no need to keep tables of opcode cycle counts etc., as the CPU emulation is agnostic about that. It just fetches cycles and executes them.