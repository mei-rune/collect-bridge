module("arguments_test",  package.seeall)


function wrap( ... )
	return {...}
end
function get(params)
	r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, r13, r14, r15, r16, r17, r18, r19, r20 = mj.invoke_native("arguments_test", 1, nil, 2, nil, 3, nil, 4, nil, 5, nil, 6, nil, 7, nil, 8, nil, 9, nil)
	mj.log(mj.SYSTEM, "R1=" .. r1)
	mj.log(mj.SYSTEM, "R2=" ..(r2 or "nil"))
	mj.log(mj.SYSTEM, "R3=" ..r3)
	mj.log(mj.SYSTEM, "R4=" ..(r4 or "nil"))
	mj.log(mj.SYSTEM, "R5=" ..r5)
	mj.log(mj.SYSTEM, "R6=" ..(r6 or "nil"))
	mj.log(mj.SYSTEM, "R7=" ..r7)
	mj.log(mj.SYSTEM, "R8=" ..(r8 or "nil"))
	mj.log(mj.SYSTEM, "R9=" ..r9)
	mj.log(mj.SYSTEM, "R10=" ..(r10 or "nil"))
	mj.log(mj.SYSTEM, "R11=" ..r11)
	mj.log(mj.SYSTEM, "R2=" ..(r12 or "nil"))
	mj.log(mj.SYSTEM, "R3=" ..r13)
	mj.log(mj.SYSTEM, "R14=" ..(r14 or "nil"))
	mj.log(mj.SYSTEM, "R15=" ..r15)
	mj.log(mj.SYSTEM, "R16=" ..(r16 or "nil"))
	mj.log(mj.SYSTEM, "R17=" ..r17)
	mj.log(mj.SYSTEM, "R18=" ..(r18 or "nil"))
	mj.log(mj.SYSTEM, "R19=" ..r19)
	mj.log(mj.SYSTEM, "R20=" ..(r20 or "nil"))


	ff = wrap(mj.invoke_native("arguments_test", 1, nil, 2, nil, 3, nil, 4, nil, 5, nil, 6, nil, 7, nil, 8, nil, 9, nil))
	for idx, value in pairs(ff) do
		mj.log(mj.SYSTEM, "ff=" .. (value or "nil"))
	end
	return {value= "ok"}, nil
end
