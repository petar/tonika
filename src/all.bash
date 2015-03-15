#!/bin/sh
cd ${TONIKAROOT}/src/pkg && gmake clean && gmake && cd ${TONIKAROOT}/src/cmd && gmake clean && gmake
