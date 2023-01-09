package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/iovisor/gobpf/bcc"
)

const source string = `
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>
#include <linux/fs.h>

#define MAX_SYSCALLS 512

struct data_t {
    u64 id;
    u64 ts;
    char comm[TASK_COMM_LEN];
    int nr;
    u64 args[6];
    int ret;
};

BPF_HASH(syscalls, u64, struct data_t);
BPF_HASH(enter_ts, u32, u64);

int trace_sys_enter(struct pt_regs *ctx) {
	u64 id = bpf_get_current_pid_tgid();
	u64 ts = bpf_ktime_get_ns();
	u32 pid = bpf_get_current_pid_tgid();

	enter_ts.update(&pid, &ts);

	return 0;
}

int trace_sys_exit(struct pt_regs *ctx) {
	struct data_t data = {};
	u64 *tsp, id = bpf_get_current_pid_tgid();
	u32 pid = id;

	tsp = enter_ts.lookup(&pid);
	if (tsp == 0)
		return 0;

	data.id = id;
	data.ts = bpf_ktime_get_ns();
	bpf_get_current_comm(&data.comm, sizeof(data.comm));
	data.nr = ctx->ax;
	data.ret = PT_REGS_RC(ctx);
	data.args[0] = PT_REGS_PARM1(ctx);
	data.args[1] = PT_REGS_PARM2(ctx);
	data.args[2] = PT_REGS_PARM3(ctx);
	data.args[3] = PT_REGS_PARM4(ctx);
	data.args[4] = PT_REGS_PARM5(ctx);
	data.args[5] = PT_REGS_PARM6(ctx);

	syscalls.update(&id, &data);

	enter_ts.delete(&pid);

	return 0;
}
`

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "PID")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid PID")
		os.Exit(1)
	}

	module := bcc.NewModule(source, []string{})
	defer module.Close()

	syscallNr, err := module.LoadKprobe("trace_sys_enter")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	syscallRet, err := module.LoadKprobe("trace_sys_exit")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = module.AttachKprobe("sys_enter", syscallNr, pid)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = module.AttachKprobe("sys_exit", syscallRet, pid)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	table := bcc.NewTable(module.TableId("syscalls"), module)

	var syscalls []syscallData

	for {
		syscallIter := table.Iter()
		syscalls = syscalls[:0]
		for {
			key, data := syscallIter.Key(), syscallIter.Leaf()

			comm := strings.TrimRight(string(data), "\x00")
			name, err := strconv.Atoi(string(key))
			if err != nil {
				name = 0
			}

			syscalls = append(syscalls, syscallData{
				TS:    uint64(data.ts),
				PID:   name,
				Comm:  comm,
				NR:    int(data.nr),
				Args:  data.args,
				Ret:   int(data.ret),
				Stack: getStack(pid, data.ts),
			})
		}

		for _, s := range syscalls {
			fmt.Printf("%d %s %d %d %d %d %d %d\n",
				s.TS, s.Comm, s.PID, s.NR, s.Args[0], s.Args[1],
				s.Args[2], s.Ret)
			fmt.Println(strings.Join(s.Stack, "\n"))
		}

		time.Sleep(time.Second)
	}
}

type syscallData struct {
	TS    uint64
	PID   int
	Comm  string
	NR    int
	Args  [6]uint64
	Ret   int
	Stack []string
}

func getStack(pid int, ts uint64) []string {
	var (
		cmd      []byte
		stack    []string
		buff     []byte
		ptrSize  = unsafe.Sizeof(uintptr(0))
		bitness  = 32
		comm     = "comm"
		exe      = "exe"
		f        *os.File
		maps     *os.File
		offset   uint64
		pc       uintptr
		pcStr    string
		prevPc   uintptr
		readByte = func(m map[uintptr]uintptr) {
			for i := uintptr(0); i < ptrSize; i++ {
				pc += uintptr(buff[i]) << (8 * i)
			}
			m[pc] = uintptr(offset)
		}
		sl      = make([]uintptr, 0, 32)
		symbols = make(map[uintptr]uintptr)
		err     error
	)

	for i := 0; ; i++ {
		cmd = append(cmd[:0], fmt.Sprintf("/proc/%d/task/%d/%s", pid, pid, exe)...)
		if f, err = os.Open(string(cmd)); err == nil {
			break
		}
		cmd = append(cmd[:0], fmt.Sprintf("/proc/%d/%s", pid, exe)...)
		if f, err = os.Open(string(cmd)); err == nil {
			break
		}
		if i == 0 {
			cmd = append(cmd[:0], fmt.Sprintf("/proc/%d/%s", pid, comm)...)
			if f, err = os.Open(string(cmd)); err == nil {
				bitness = 8
				break
			}
		}
		if i > 1 {
			break
		}
		cmd = cmd[:0]
	}
	if err != nil {
		return nil
	}
	defer f.Close()

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(0x0000dc01), uintptr(unsafe.Pointer(&offset))); errno != 0 {
		return nil
	}

	cmd = append(cmd[:0], fmt.Sprintf("/proc/%d/task/%d/%s", pid, pid, "maps")...)
	maps, err = os.Open(string(cmd))
	if err != nil {
		return nil
	}
	defer maps.Close()
	for {
		_, err = fmt.Fscanf(maps, "%x-%x", &pc, &prevPc)
		if err != nil {
			break
		}
		prevPc = pc
		pc = 0
		buff = make([]byte, ptrSize)
		_, err = fmt.Fscanf(maps, "%s %*s %*s %*s %*d %s", &pcStr, &pcStr)
		if err != nil {
			break
		}
		if pcStr == "00" {
			pc = prevPc
			readByte(symbols)
			continue
		}
		if _, err = fmt.Sscanf(pcStr, "%x", &pc); err != nil {
			break
		}
		if pc < prevPc {
			pc = prevPc
			readByte(symbols)
			continue
		}
		pc += prevPc
		for _, b := range []byte(pcStr) {
			if b >= 'a' && b <= 'f' {
				bitness = 64
				break
			}
		}
		readByte(symbols)
		pc = 0
		for {
			if _, err = fmt.Fscanf(maps, "%x", &pc); err != nil {
				break
			}
			pc += prevPc
			readByte(symbols)
			pc = 0
		}
	}

	if err != nil && err != io.EOF {
		return nil
	}

	if errno := syscall.PtraceSetOptions(pid, syscall.PTRACE_O_TRACESYSGOOD); errno != nil {
		return nil
	}
	for {
		if errno := syscall.PtraceSyscall(pid, 0); errno != nil {
			break
		}
		_, err = syscall.Wait4(pid, nil, 0, nil)
		if err != nil {
			break
		}
		var regs syscall.PtraceRegs
		if errno := syscall.PtraceGetRegs(pid, &regs); errno != nil {
			break
		}
		if bitness == 32 {
			pc = uintptr(regs.Rip)
			prevPc = uintptr(regs.Rbp)
			for i := 0; i < 16; i++ {
				pc += uintptr(buff[i]) << (8 * i)
			}
			if offset, ok := symbols[pc]; ok {
				if _, err = f.Seek(int64(offset), 0); err != nil {
					break
				}
				if _, err = f.Read(buff[:bitness/8]); err != nil {
					break
				}
				if _, err = fmt.Sscanf(string(buff[:bitness/8]), "%x", &pc); err != nil {
					break
				}
				sl = append(sl, pc-offset)
				if pc == prevPc {
					break
				}
				pc = prevPc
				prevPc = 0
				for i := 0; i < 16; i++ {
					prevPc += uintptr(buff[i]) << (8 * i)
				}
			}
			break
		}
		pc = uintptr(regs.Rip)
		prevPc = uintptr(regs.Rbp)
		for {
			if offset, ok := symbols[pc]; ok {
				if _, err = f.Seek(int64(offset), 0); err != nil {
					break
				}
				if _, err = f.Read(buff); err != nil {
					break
				}
				if _, err = fmt.Sscanf(string(buff), "%x", &pc); err != nil {
					break
				}
				sl = append(sl, pc-offset)
				if pc == prevPc {
					break
				}
				pc = prevPc
				prevPc = 0
				for i := 0; i < 16; i++ {
					prevPc += uintptr(buff[i]) << (8 * i)
				}
			}
			break
		}
		if err != nil {
			break
		}
		if len(sl) == 0 {
			break
		}
		for i, pc := range sl {
			cmd = append(cmd[:0], fmt.Sprintf("/tmp/perf-%d.map", pid)...)
			if f, err = os.Open(string(cmd)); err == nil {
				break
			}
			if i > 0 {
				break
			}
			if _, err = syscall.PtracePeekText(pid, uintptr(pc), buff); err != nil {
				break
			}
			if buff[0] == 0xcc {
				if errno := syscall.PtraceCont(pid, 0); errno != nil {
					break
				}
				_, err = syscall.Wait4(pid, nil, 0, nil)
				if err != nil {
					break
				}
				if errno := syscall.PtraceGetRegs(pid, &regs); errno != nil {
					break
				}
				pc = uintptr(regs.Rip)
			}
			break
		}
		if err != nil {
			break
		}
		defer f.Close()
		for _, pc := range sl {
			if _, err = f.Seek(int64(pc), 0); err != nil {
				break
			}
			if _, err = f.Read(buff[:bitness/8]); err != nil {
				break
			}
			if _, err = fmt.Sscanf(string(buff[:bitness/8]), "%x", &pc); err != nil {
				break
			}
			stack = append(stack, fmt.Sprintf("%x", pc))
			if pc == 0 {
				break
			}
		}
		break
	}
	if err != nil {
		return nil
	}
	return stack
}
